package driver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/josenarvaezp/displ/mocks"
	"github.com/josenarvaezp/displ/pkg/lambdas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StartCoordinator_HappyPath2(t *testing.T) {
	ctx := context.Background()
	jobId := uuid.New()
	numMappers := 10
	numReducers := 2
	functionArn := "arn:aws:lambda:eu-west-2:000000000000:function:test-name"
	request := &lambdas.CoordinatorInput{
		JobID:      jobId,
		NumMappers: numMappers,
		NumQueues:  numReducers,
	}

	// expected payload
	requestPayload, err := json.Marshal(request)
	require.Nil(t, err)

	expectedInvokeInput := &lambda.InvokeInput{
		FunctionName:   aws.String(functionArn),
		Payload:        requestPayload,
		InvocationType: lambdaTypes.InvocationTypeEvent,
	}

	expectedResult := &lambda.InvokeOutput{
		StatusCode: 202,
	}

	lambdaMock := new(mocks.FaasAPI)
	lambdaMock.On("Invoke", ctx, expectedInvokeInput).Return(expectedResult, nil)

	jobDriver := Driver{
		JobID: jobId,
		Config: config.Config{
			Region:    "eu-west-2",
			AccountID: "000000000000",
		},
		BuildData: &generators.BuildData{
			CoordinatorData: &generators.CoordinatorData{
				ImageName: "test-name",
			},
		},
		FaasAPI: lambdaMock,
	}

	err = jobDriver.StartCoordinator(ctx, numMappers, numReducers)
	assert.Nil(t, err)
}

var jobConfig config.Config
var jobBuildData *generators.BuildData
var expectedInvokeInput string
var expectedResult string

func Test_StartCoordinator_HappyPath(t *testing.T) {
	// set input
	ctx := context.Background()
	numMappers := 10
	numReducers := 2

	// set mock
	lambdaMock := new(mocks.FaasAPI)
	lambdaMock.On("Invoke", ctx, expectedInvokeInput).
		Return(expectedResult, nil)

	// init driver
	jobDriver := Driver{
		JobID:     uuid.New(),
		Config:    jobConfig,
		BuildData: jobBuildData,
		FaasAPI:   lambdaMock,
	}

	err := jobDriver.StartCoordinator(ctx, numMappers, numReducers)
	assert.Nil(t, err)
}
