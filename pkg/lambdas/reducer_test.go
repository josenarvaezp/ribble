package lambdas_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/mocks"
	"github.com/josenarvaezp/displ/pkg/lambdas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getExpectedMessage(t *testing.T) *string {
	meta := &lambdas.QueueMetadata{
		MapID:      uuid.New().String(),
		NumBatches: 5,
	}
	p, err := json.Marshal(meta)
	require.Nil(t, err)

	metaJSONString := string(p)
	return &metaJSONString
}

func Test_GetNumberOfBatchesToProcess_HappyPath(t *testing.T) {
	ctx := context.Background()
	jobID := uuid.New()
	queueUrl := fmt.Sprintf(
		"https://sqs.eu-west-2.amazonaws.com/000000000000/%s-1-meta",
		jobID.String(),
	)

	// set expectations
	expectedInput := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5,
	}

	expectedOutput1 := &sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				Body: getExpectedMessage(t),
			},
		},
	}

	expectedOutput2 := &sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				Body: getExpectedMessage(t),
			},
		},
	}

	// create mock
	sqsMock := new(mocks.QueuesAPI)
	// simulate a repeated message - one should be ignored
	sqsMock.On("ReceiveMessage", ctx, expectedInput).Return(expectedOutput1, nil).Twice()
	// simulate another message
	sqsMock.On("ReceiveMessage", ctx, expectedInput).Return(expectedOutput2, nil).Once()

	reducer := lambdas.Reducer{
		JobID:          jobID,
		Region:         "eu-west-2",
		AccountID:      "000000000000",
		Local:          false,
		NumMappers:     2,
		QueuePartition: 1,
		QueuesAPI:      sqsMock,
	}

	numBatches, err := reducer.GetNumberOfBatchesToProcess(ctx)
	require.Nil(t, err)
	// assert numBatches is 10 as 2 mappers were used
	// each indicating they sent 5 messages
	assert.Equal(t, 10, *numBatches)
}
