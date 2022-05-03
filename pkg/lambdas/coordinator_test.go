package lambdas_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/mocks"
	"github.com/josenarvaezp/displ/pkg/lambdas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UpdateCoordinator_HappyPath(t *testing.T) {

	coordinator := &lambdas.Coordinator{}

	jobID := uuid.New()
	input := lambdas.CoordinatorInput{
		JobID:      jobID,
		NumMappers: 4,
		NumQueues:  2,
	}

	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: "arn:aws:lambda:eu-west-2:000000000000:function:coordinator-name",
	})

	err := coordinator.UpdateCoordinatorWithRequest(ctx, input)
	assert.Nil(t, err)

	assert.Equal(t, jobID, coordinator.JobID)
	assert.Equal(t, "000000000000", coordinator.AccountID)
	assert.Equal(t, int64(4), coordinator.NumMappers)
	assert.Equal(t, int64(2), coordinator.NumQueues)
}

func Test_InvokeReducers_HappyPath(t *testing.T) {
	ctx := context.Background()
	filename := "reducers-invoked"
	reducerName := "reducerName"
	functionARN := "arn:aws:lambda:eu-west-2:000000000000:function:reducer"

	jobID := uuid.New()

	expectedGetObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(jobID.String()),
		Key:    aws.String(filename),
	}

	expectedUploadInput := &s3.PutObjectInput{
		Bucket: aws.String(jobID.String()),
		Key:    aws.String(filename),
		Body:   bytes.NewReader([]byte{}),
	}

	s3Mock := new(mocks.ObjectStoreAPI)
	s3Mock.On("GetObject", ctx, expectedGetObjectInput).Return(&s3.GetObjectOutput{}, nil)
	s3Mock.On("Upload", ctx, expectedUploadInput).Return(nil)

	reducerInput := lambdas.ReducerInput{
		JobID:          jobID,
		ReducerID:      uuid.New(),
		QueuePartition: 0,
		NumMappers:     5,
	}
	requestPayload, err := json.Marshal(reducerInput)
	require.Nil(t, err)

	expectedInvokeInput1 := &lambda.InvokeInput{
		FunctionName:   aws.String(functionARN),
		Payload:        requestPayload,
		InvocationType: types.InvocationTypeEvent,
	}

	reducerInput2 := lambdas.ReducerInput{
		JobID:          jobID,
		ReducerID:      uuid.New(),
		QueuePartition: 1,
		NumMappers:     5,
	}
	requestPayload2, err := json.Marshal(reducerInput2)
	require.Nil(t, err)

	expectedInvokeInput2 := &lambda.InvokeInput{
		FunctionName:   aws.String(functionARN),
		Payload:        requestPayload2,
		InvocationType: types.InvocationTypeEvent,
	}

	expectedInkokeOutput := &lambda.InvokeOutput{
		StatusCode: int32(202),
	}

	lambdaMock := new(mocks.FaasAPI)
	lambdaMock.On("Invoke", ctx, expectedInvokeInput1).Return(expectedInkokeOutput, nil).Once()
	lambdaMock.On("Invoke", ctx, expectedInvokeInput2).Return(expectedInkokeOutput, nil).Once()

	coordinator := &lambdas.Coordinator{
		JobID:          jobID,
		Region:         "eu-west-2",
		AccountID:      "000000000000",
		NumQueues:      2,
		NumMappers:     5,
		ObjectStoreAPI: s3Mock,
		FaasAPI:        lambdaMock,
	}

	err = coordinator.InvokeReducers(ctx, reducerName)
	assert.Nil(t, err)
}
