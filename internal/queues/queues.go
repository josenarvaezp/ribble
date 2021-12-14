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
}
