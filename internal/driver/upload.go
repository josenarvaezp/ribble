package driver

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrTypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
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

// UploadLambdaFunctions upploads the map, coordinator and
// reducer images needed for the job and creates the lambda function
func (d *Driver) UploadLambdaFunctions(ctx context.Context, dqlARN *string) error {
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
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) CreateLambdaFunction(ctx context.Context, imageName, imageTag string, imageURI *string, lambdaDlqArn *string) error {
	functionDescription := fmt.Sprintf("Ribble function for %s", imageName)
	functionTimeout := int32(90)
	ribbleRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/ribble", d.Config.AccountID)

	_, err := d.FaasAPI.CreateFunction(ctx, &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: imageURI,
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
	})

	return err
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
