package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrTypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/josenarvaezp/displ/pkg/lambdas"
)

const (
	scriptToUploadImages        = "./build/upload_images.sh"
	scriptToUploadImagesLocally = "./build/upload_images_locally.sh"
)

// CreateRepo creates a repository in ECR to upload an image
func (d *Driver) CreateRepo(ctx context.Context, repoName string) (*string, error) {
	// create repo
	params := &ecr.CreateRepositoryInput{
		RepositoryName: &repoName,
		RegistryId:     &d.Config.AccountID,
	}
	res, err := d.ImageRepoAPI.CreateRepository(ctx, params)
	if err != nil {
		return nil, err
	}

	return res.Repository.RepositoryUri, nil
}

// UploadMapper upploads a mapper image to ECR
func (d *Driver) UploadImage(ctx context.Context, imageName, imageTag string) (*string, error) {
	repoURI, err := d.CreateRepo(ctx, imageName)
	if err != nil {
		return nil, err
	}

	if d.Config.Local {
		// tag and push image to localstack
		if _, err := exec.Command(
			scriptToUploadImagesLocally,
			imageName,
			imageTag,
			*repoURI,
		).Output(); err != nil {
			return nil, err
		}
	} else {
		// tag and push image to AWS
		if _, err := exec.Command(
			scriptToUploadImages,
			imageName,
			imageTag,
			d.Config.AccountID,
			d.Config.Region,
		).Output(); err != nil {
			return nil, err
		}
	}

	return repoURI, nil
}

func (d *Driver) UploadLambdaFunctionsLocal() error {
	// to run lambda in localstack we need to create
	// the functions as zip rather than containers

	// create mapper lambda function
	if _, err := exec.Command(
		scriptToUploadImagesLocally,
		d.BuildData.MapperData.GeneratedFile,
		d.BuildData.MapperData.Function,
		d.Config.Region,
		d.BuildData.MapperData.ImageName,
	).Output(); err != nil {
		return err
	}

	// // create coordinator lambda function
	if _, err := exec.Command(
		scriptToUploadImagesLocally,
		d.BuildData.CoordinatorData.GeneratedFile,
		d.BuildData.CoordinatorData.Function,
		d.Config.Region,
		d.BuildData.CoordinatorData.ImageName,
	).Output(); err != nil {
		return err
	}

	// create reducer lambda function
	for _, reducer := range d.BuildData.ReducerData {
		if _, err := exec.Command(
			scriptToUploadImagesLocally,
			reducer.GeneratedFile,
			reducer.ReducerName,
			d.Config.Region,
			reducer.ImageName,
		).Output(); err != nil {
			return err
		}
	}

	return nil
}

// UploadLambdaFunctions upploads the map, coordinator and
// reducer images needed for the job and creates the lambda function
func (d *Driver) UploadLambdaFunctions(ctx context.Context, dqlARN *string) error {
	if d.Config.Local {
		// upload images to localstack as zip files
		err := d.UploadLambdaFunctionsLocal()
		if err != nil {
			return err
		}

		return nil
	}

	// upload mapper image
	mapperURI, err := d.UploadImage(
		ctx,
		d.BuildData.MapperData.ImageName,
		d.BuildData.MapperData.ImageTag,
	)
	if err != nil {
		return err
	}

	// create lambda mapper function
	err = d.CreateLambdaFunction(
		ctx,
		d.BuildData.MapperData.ImageName,
		d.BuildData.MapperData.ImageTag,
		mapperURI,
		dqlARN,
		128,
	)
	if err != nil {
		return err
	}

	// upload coordinator image
	coordinatorURI, err := d.UploadImage(
		ctx,
		d.BuildData.CoordinatorData.ImageName,
		d.BuildData.CoordinatorData.ImageTag,
	)
	if err != nil {
		return err
	}

	// create lambda coordinator function
	err = d.CreateLambdaFunction(
		ctx,
		d.BuildData.CoordinatorData.ImageName,
		d.BuildData.CoordinatorData.ImageTag,
		coordinatorURI,
		dqlARN,
		512,
	)
	if err != nil {
		return err
	}

	// create reducers
	for _, reducer := range d.BuildData.ReducerData {
		// upload reducer image
		currentURI, err := d.UploadImage(
			ctx,
			reducer.ImageName,
			reducer.ImageTag,
		)
		if err != nil {
			return err
		}

		// create lambda reducer function
		err = d.CreateLambdaFunction(
			ctx,
			reducer.ImageName,
			reducer.ImageTag,
			currentURI,
			dqlARN,
			512,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) CreateLambdaFunction(
	ctx context.Context,
	imageName,
	imageTag string,
	imageURI *string,
	lambdaDlqArn *string,
	memory int32,
) error {
	functionDescription := fmt.Sprintf("Ribble function for %s", imageName)
	functionTimeout := int32(900)
	ribbleRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/ribble", d.Config.AccountID)
	imageURIWithTag := fmt.Sprintf("%s:%s", *imageURI, imageTag)

	_, err := d.FaasAPI.CreateFunction(ctx, &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: &imageURIWithTag,
		},
		FunctionName: &imageName,
		Role:         &ribbleRoleArn,
		DeadLetterConfig: &types.DeadLetterConfig{
			TargetArn: lambdaDlqArn,
		},
		Description: &functionDescription,
		PackageType: types.PackageTypeImage,
		Publish:     true,
		Timeout:     &functionTimeout,
		MemorySize:  aws.Int32(memory),
	})

	return err
}

// CreateLogs creates a log group and creates a log stream
// for the job
func (d *Driver) CreateLogsInfra(ctx context.Context) error {
	// create log group
	logGroupName := fmt.Sprintf("%s-log-group", d.JobID.String())
	_, err := d.LogsAPI.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &logGroupName,
	})
	if err != nil {
		return err
	}

	// create log stream
	logStreamName := fmt.Sprintf("%s-log-stream", d.JobID.String())
	_, err = d.LogsAPI.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &logStreamName,
	})
	if err != nil {
		return err
	}

	return nil
}

// WriteMappings writes the mappings to s3 so that the coordinator can read them
func (d *Driver) WriteMappings(ctx context.Context, mappings []*lambdas.Mapping) error {
	// encode map to JSON
	p, err := json.Marshal(mappings)
	if err != nil {
		return err
	}

	// use uploader manager to write file to S3
	jsonContentType := "application/json"
	bucket := d.JobID.String()
	input := &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           aws.String("mappings"),
		Body:          bytes.NewReader(p),
		ContentType:   &jsonContentType,
		ContentLength: int64(len(p)),
	}
	_, err = d.UploaderAPI.Upload(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

// repoAlreadyExists checks if the ecr repo being created already exists
func repoAlreadyExists(err error) bool {
	var alreadyExists *ecrTypes.RepositoryAlreadyExistsException
	return errors.As(err, &alreadyExists)
}

// TODO:
// func (d *Driver) configure(ctx context.Context, functionName *string, reserveConcurrency int32, reserveConcurrency ) {
// 	d.FaasAPI.PutFunctionConcurrency(ctx, &lambda.PutFunctionConcurrencyInput{
// 		FunctionName:                 functionName,
// 		ReservedConcurrentExecutions: &reserveConcurrency,
// 	})

// 	d.FaasAPI.PutProvisionedConcurrencyConfig(ctx, &lambda.PutProvisionedConcurrencyConfigInput{
// 		FunctionName: functionName,
// 		ProvisionedConcurrentExecutions: "",
// 		Qualifier: "",
// 	})
// }
