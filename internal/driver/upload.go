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
	scriptToUploadImages = "./build/upload_images.sh"
)

// CreateRepo creates a repository in ECR to upload an image
func (d *Driver) CreateRepo(ctx context.Context, repoName string) error {
	// create repo
	params := &ecr.CreateRepositoryInput{
		RepositoryName: &repoName,
		RegistryId:     &d.Config.AccountID,
	}
	_, err := d.ImageRepoAPI.CreateRepository(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// UploadMapper upploads a mapper image to ECR
func (d *Driver) UploadMapper(ctx context.Context) error {
	err := d.CreateRepo(ctx, d.BuildData.MapperData.ImageName)
	if err != nil {
		return err
	}

	// tag and push image
	_, err = exec.Command(
		scriptToUploadImages,
		d.BuildData.MapperData.ImageName,
		d.BuildData.MapperData.ImageTag,
		d.Config.AccountID,
		d.Config.Region,
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// UploadCoordinator upploads a coordinator image to ECR
func (d *Driver) UploadCoordinator(ctx context.Context) error {
	err := d.CreateRepo(ctx, d.BuildData.CoordinatorData.ImageName)
	if err != nil {
		return err
	}

	// tag and push image
	_, err = exec.Command(
		scriptToUploadImages,
		d.BuildData.CoordinatorData.ImageName,
		d.BuildData.CoordinatorData.ImageTag,
		d.Config.AccountID,
		d.Config.Region,
	).Output()
	if err != nil {
		return err
	}

	return nil
}

// UploadAggregators upploads the aggregator images to ECR
func (d *Driver) UploadAggregators(ctx context.Context) error {
	for _, reducer := range d.BuildData.ReducerData {
		err := d.CreateRepo(ctx, reducer.ReducerName)
		if err != nil {
			// only ignore error if repo exists already
			if !repoAlreadyExists(err) {
				return err
			}
			return err
		}

		// tag and push image
		_, err = exec.Command(
			scriptToUploadImages,
			reducer.ReducerName,
			"latest",
			d.Config.AccountID,
			d.Config.Region,
		).Output()
		if err != nil {
			return err
		}
	}

	return nil
}

// UploadJobImages upploads the map, coordinator and
// aggregator images needed for the job
func (d *Driver) UploadJobImages(ctx context.Context) error {
	err := d.UploadMapper(ctx)
	if err != nil {
		return err
	}

	err = d.UploadCoordinator(ctx)
	if err != nil {
		return err
	}

	err = d.UploadAggregators(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) CreateMapperLambdaFunction(ctx context.Context, lambdaDlqArn *string) error {
	imageURI := fmt.Sprintf(
		"%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		d.Config.AccountID,
		d.Config.Region,
		d.BuildData.MapperData.ImageName,
		d.BuildData.MapperData.ImageTag,
	)
	functionDescription := fmt.Sprintf("Ribble function for %s", d.BuildData.MapperData.Function)
	functionTimeout := int32(900)
	ribbleRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/ribble", d.Config.AccountID)
	functionName := fmt.Sprintf("%s_%s", d.BuildData.MapperData.Function, d.JobID.String())

	_, err := d.FaasAPI.CreateFunction(ctx, &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: &imageURI,
		},
		FunctionName: &functionName,
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

func (d *Driver) CreateCoordinatorLambdaFunction(ctx context.Context, lambdaDlqArn *string) error {
	imageURI := fmt.Sprintf(
		"%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		d.Config.AccountID,
		d.Config.Region,
		d.BuildData.CoordinatorData.ImageName,
		d.BuildData.CoordinatorData.ImageTag,
	)
	functionDescription := fmt.Sprintf("Ribble function for %s", d.BuildData.CoordinatorData.Function)
	functionTimeout := int32(900)
	ribbleRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/ribble", d.Config.AccountID)
	functionName := fmt.Sprintf("%s_%s", d.BuildData.CoordinatorData.Function, d.JobID.String())

	_, err := d.FaasAPI.CreateFunction(ctx, &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: &imageURI,
		},
		FunctionName: &functionName,
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

func (d *Driver) CreateAggregatorLambdaFunctions(ctx context.Context, queueARN *string) error {
	ribbleRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/ribble", d.Config.AccountID)
	functionTimeout := int32(900)

	for _, reducer := range d.BuildData.ReducerData {
		imageURI := fmt.Sprintf(
			"%s.dkr.ecr.%s.amazonaws.com/%s:%s",
			d.Config.AccountID,
			d.Config.Region,
			reducer.ImageName,
			"latest",
		)
		functionDescription := fmt.Sprintf("Ribble function for %s", reducer.ReducerName)

		_, err := d.FaasAPI.CreateFunction(ctx, &lambda.CreateFunctionInput{
			Code: &types.FunctionCode{
				ImageUri: &imageURI,
			},
			FunctionName: &reducer.ReducerName,
			Role:         &ribbleRoleArn,
			DeadLetterConfig: &types.DeadLetterConfig{
				TargetArn: queueARN,
			},
			Description: &functionDescription,
			PackageType: types.PackageTypeImage,
			Publish:     true,
			Timeout:     &functionTimeout,
		})
		if err != nil {
			return err
		}
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
