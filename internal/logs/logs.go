package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// LogsAPI is an interface used to mock API calls made to the aws cloudwatch service
type LogsAPI interface {
	CreateLogGroup(
		ctx context.Context,
		params *cloudwatchlogs.CreateLogGroupInput,
		optFns ...func(*cloudwatchlogs.Options),
	) (*cloudwatchlogs.CreateLogGroupOutput, error)
	CreateLogStream(
		ctx context.Context,
		params *cloudwatchlogs.CreateLogStreamInput,
		optFns ...func(*cloudwatchlogs.Options),
	) (*cloudwatchlogs.CreateLogStreamOutput, error)
	PutLogEvents(
		ctx context.Context,
		params *cloudwatchlogs.PutLogEventsInput,
		optFns ...func(*cloudwatchlogs.Options),
	) (*cloudwatchlogs.PutLogEventsOutput, error)
}
