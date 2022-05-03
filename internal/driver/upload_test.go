package driver

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrTypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRepo_HappyPath(t *testing.T) {
	ctx := context.Background()
	jobId := uuid.New()

	repoName := "test-repo-name"
	accountID := "000000000000"
	repoURI := "000000000000.dkr.ecr.eu-west-2.amazonaws.com/test-repo-name"

	expectedInput := &ecr.CreateRepositoryInput{
		RepositoryName: &repoName,
		RegistryId:     &accountID,
	}

	expectedResult := &ecr.CreateRepositoryOutput{
		Repository: &ecrTypes.Repository{
			RepositoryUri: &repoURI,
		},
	}

	ecrMock := new(mocks.ImageRepoAPI)
	ecrMock.On("CreateRepository", ctx, expectedInput).Return(expectedResult, nil)

	jobDriver := Driver{
		JobID: jobId,
		Config: config.Config{
			Region:    "eu-west-2",
			AccountID: "000000000000",
		},
		ImageRepoAPI: ecrMock,
	}

	resUri, err := jobDriver.CreateRepo(ctx, repoName)
	assert.Nil(t, err)
	assert.Equal(t, repoURI, *resUri)
}

func Test_CreateRepo_UnhappyPath(t *testing.T) {
	ctx := context.Background()
	jobId := uuid.New()

	repoName := "test-repo-name"
	accountID := "000000000000"

	expectedInput := &ecr.CreateRepositoryInput{
		RepositoryName: &repoName,
		RegistryId:     &accountID,
	}

	ecrMock := new(mocks.ImageRepoAPI)
	ecrMock.On("CreateRepository", ctx, expectedInput).Return(nil, errors.New("mock error"))

	jobDriver := Driver{
		JobID: jobId,
		Config: config.Config{
			Region:    "eu-west-2",
			AccountID: "000000000000",
		},
		ImageRepoAPI: ecrMock,
	}

	resUri, err := jobDriver.CreateRepo(ctx, repoName)
	assert.Nil(t, resUri)
	assert.EqualError(t, err, "mock error")

}

func Test_CreateLambdaFunction_HappyPath(t *testing.T) {
	ctx := context.Background()
	jobId := uuid.New()

	imageName := "test-name"
	imageTag := "latest"
	imageURI := "000000000000.dkr.ecr.eu-west-2.amazonaws.com/test-name"
	lambdaDlqArn := "test-dlq"
	memory := int32(128)

	functionDescription := "Ribble function for test-name"
	functionTimeout := int32(900)
	ribbleRoleArn := "arn:aws:iam::000000000000:role/ribble"
	imageURIWithTag := "000000000000.dkr.ecr.eu-west-2.amazonaws.com/test-name:latest"

	expectedInput := &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: &imageURIWithTag,
		},
		FunctionName: &imageName,
		Role:         &ribbleRoleArn,
		DeadLetterConfig: &types.DeadLetterConfig{
			TargetArn: &lambdaDlqArn,
		},
		Description: &functionDescription,
		PackageType: types.PackageTypeImage,
		Publish:     true,
		Timeout:     &functionTimeout,
		MemorySize:  &memory,
	}

	lambdaMock := new(mocks.FaasAPI)
	lambdaMock.On("CreateFunction", ctx, expectedInput).Return(&lambda.CreateFunctionOutput{}, nil)

	jobDriver := Driver{
		JobID: jobId,
		Config: config.Config{
			Region:    "eu-west-2",
			AccountID: "000000000000",
		},
		FaasAPI: lambdaMock,
	}

	err := jobDriver.CreateLambdaFunction(
		ctx,
		imageName,
		imageTag,
		&imageURI,
		&lambdaDlqArn,
		memory,
	)
	assert.Nil(t, err)
}

func Test_CreateLambdaFunction_UnhappyPath(t *testing.T) {
	ctx := context.Background()
	jobId := uuid.New()

	imageName := "test-name"
	imageTag := "latest"
	imageURI := "000000000000.dkr.ecr.eu-west-2.amazonaws.com/test-name"
	lambdaDlqArn := "test-dlq"
	memory := int32(128993293) // invalid

	functionDescription := "Ribble function for test-name"
	functionTimeout := int32(900)
	ribbleRoleArn := "arn:aws:iam::000000000000:role/ribble"
	imageURIWithTag := "000000000000.dkr.ecr.eu-west-2.amazonaws.com/test-name:latest"

	expectedInput := &lambda.CreateFunctionInput{
		Code: &types.FunctionCode{
			ImageUri: &imageURIWithTag,
		},
		FunctionName: &imageName,
		Role:         &ribbleRoleArn,
		DeadLetterConfig: &types.DeadLetterConfig{
			TargetArn: &lambdaDlqArn,
		},
		Description: &functionDescription,
		PackageType: types.PackageTypeImage,
		Publish:     true,
		Timeout:     &functionTimeout,
		MemorySize:  &memory,
	}

	lambdaMock := new(mocks.FaasAPI)
	lambdaMock.On("CreateFunction", ctx, expectedInput).Return(nil, errors.New("mock invalid memory"))

	jobDriver := Driver{
		JobID: jobId,
		Config: config.Config{
			Region:    "eu-west-2",
			AccountID: "000000000000",
		},
		FaasAPI: lambdaMock,
	}

	err := jobDriver.CreateLambdaFunction(
		ctx,
		imageName,
		imageTag,
		&imageURI,
		&lambdaDlqArn,
		memory,
	)
	assert.EqualError(t, err, "mock invalid memory")
}
