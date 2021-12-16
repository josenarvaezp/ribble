package queues

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// QueuesAPI is an interface used to mock API calls made to the aws SQS service
type QueuesAPI interface {
	CreateQueue(
		ctx context.Context,
		params *sqs.CreateQueueInput,
		optFns ...func(*sqs.Options),
	) (*sqs.CreateQueueOutput, error)
	GetQueueAttributes(
		ctx context.Context,
		params *sqs.GetQueueAttributesInput,
		optFns ...func(*sqs.Options),
	) (*sqs.GetQueueAttributesOutput, error)
	SendMessageBatch(
		ctx context.Context,
		params *sqs.SendMessageBatchInput,
		optFns ...func(*sqs.Options),
	) (*sqs.SendMessageBatchOutput, error)
	GetQueueUrl(
		ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options),
	) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessage(
		ctx context.Context,
		params *sqs.ReceiveMessageInput,
		optFns ...func(*sqs.Options),
	) (*sqs.ReceiveMessageOutput, error)
}
