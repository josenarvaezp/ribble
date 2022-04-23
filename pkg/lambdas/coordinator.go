package lambdas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudWatchTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/faas"
	"github.com/josenarvaezp/displ/internal/logs"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/queues"
)

const (
	CoordinatorName = "displ-coordinator" // TODO: name of the function or ARN
)

// CoordinatorInput is the input the coordinator lambda receives
type CoordinatorInput struct {
	JobID        uuid.UUID `json:"jobID"`
	NumMappers   int       `json:"numMappers"`
	NumQueues    int       `json:"numQueues"`
	FunctionName string    `json:"functionName"`
}

// CoordinatorAPI is an interface deining the functions available to the coordinator
type CoordinatorAPI interface {
	AreMappersDone(ctx context.Context) (bool, error)
	InvokeReducers(ctx context.Context) error
}

// Coordinator is an interface that implements CoordinatorAPI
type Coordinator struct {
	JobID uuid.UUID
	// clients
	QueuesAPI      queues.QueuesAPI
	FaasAPI        faas.FaasAPI
	DownloaderAPI  objectstore.ManagerDownloaderAPI
	UploaderAPI    objectstore.ManagerUploaderAPI
	LogsAPI        logs.LogsAPI
	ObjectStoreAPI objectstore.ObjectStoreAPI
	// metadata
	Region     string
	AccountID  string
	NumMappers int64
	NumQueues  int64
	local      bool
}

// NewCoordinator initializes a new coordinator with its required clients
func NewCoordinator(
	local bool,
) (*Coordinator, error) {
	var cfg *aws.Config
	var err error

	// get region from env var
	region := os.Getenv("AWS_REGION")

	// init coordinator
	coordinator := &Coordinator{
		Region: region,
		local:  local,
	}

	// create config
	if local {
		cfg, err = config.InitLocalCfg(
			config.INTERNAL_LOCALSTACK_HOST_NAME,
			config.LOCALSTACK_PORT,
			region,
		)
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.InitCfg(region)
		if err != nil {
			return nil, err
		}
	}

	// create sqs client
	coordinator.QueuesAPI = sqs.NewFromConfig(*cfg)

	// create lambda client
	coordinator.FaasAPI = lambda.NewFromConfig(*cfg)

	// Create an S3 client using the loaded configuration
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	coordinator.DownloaderAPI = manager.NewDownloader(s3Client)
	coordinator.UploaderAPI = manager.NewUploader(s3Client)
	coordinator.ObjectStoreAPI = s3Client

	// create logs client
	coordinator.LogsAPI = cloudwatchlogs.NewFromConfig(*cfg)

	return coordinator, err
}

// UpdateCoordinatorWithRequest updates the coordinator struct with the information
// gathered from the context and request
func (c *Coordinator) UpdateCoordinatorWithRequest(ctx context.Context, request CoordinatorInput) error {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return errors.New("Error getting lambda context")
	}
	c.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	c.JobID = request.JobID
	c.NumMappers = int64(request.NumMappers)
	c.NumQueues = int64(request.NumQueues)

	return nil
}

// AreMappersDone reads events from the mapper-done queue to check
// if all mappers are done
func (c *Coordinator) AreMappersDone(ctx context.Context, nextLogToken *string) (*string, error) {
	queueName := fmt.Sprintf("%s-mappers-done", c.JobID.String())
	queueURL := GetQueueURL(queueName, c.Region, c.AccountID, c.local)
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: MaxItemsPerBatch,
	}

	// keeps a map of done mappers, this is used as the dedupe mechanism
	doneMappers := make(map[string]bool)
	doneMappersCount := 0

	// loop until all mappers are done
	for doneMappersCount < int(c.NumMappers) {
		// mappers are not done yet
		output, err := c.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return nil, err
		}

		if len(output.Messages) != 0 {
			// at least some mappers have completed
			for _, message := range output.Messages {
				// add mapper to done map
				if _, ok := doneMappers[*message.Body]; !ok {
					doneMappers[*message.Body] = true
					doneMappersCount++
				}
			}

			// break and continue processing without sleep
			break
		}

		// sleep for 1 seconds before trying to get more results
		time.Sleep(1 * time.Second)
	}

	messagesReceived := 0
	// loop until all mappers are done
	for doneMappersCount < int(c.NumMappers) {
		// mappers are not done yet
		output, err := c.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, message := range output.Messages {
			// add mapper to done map
			if _, ok := doneMappers[*message.Body]; !ok {
				doneMappers[*message.Body] = true
				doneMappersCount++
			}
		}

		// only log every 100 messages
		if messagesReceived%100 == 0 {
			if mappersCompleted, err := c.GetNumMessagesInQueue(ctx, queueURL); err == nil {
				nextLogToken, _ = c.LogEvent(
					ctx,
					fmt.Sprintf("Mappers completed: %d/%d", mappersCompleted, c.NumMappers),
					nextLogToken,
				)
			}
		}

		messagesReceived = messagesReceived + 10
	}

	return nextLogToken, nil
}

// InvokeReducers is used to invoke the reducers once all mappers are done
// there is one reducer per queue invoked
func (c *Coordinator) InvokeReducers(ctx context.Context, reducerName string) error {
	// check if reducers have been invoked
	if c.GetDoneObject(ctx, "reducers-invoked") {
		// reducers have already been invoked
		// it is likely that the coordinator crashed
		return nil
	}

	// function arn
	functionArn := fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s_%s",
		c.Region,
		c.AccountID,
		reducerName,
		c.JobID.String(),
	)

	// invoke a reducer per each queue
	for i := 0; i < int(c.NumQueues); i++ {
		// encode reducer input to json
		reducerInput := ReducerInput{
			JobID:          c.JobID,
			ReducerID:      uuid.New(),
			QueuePartition: i,
			NumMappers:     int(c.NumMappers),
		}
		requestPayload, err := json.Marshal(reducerInput)
		if err != nil {
			return err
		}

		result, err := c.FaasAPI.Invoke(
			ctx,
			&lambda.InvokeInput{
				FunctionName:   aws.String(functionArn),
				Payload:        requestPayload,
				InvocationType: types.InvocationTypeEvent,
			},
		)
		if err != nil {
			return err
		}

		// error is ignored from asynch invokation and result only holds the status code
		// check status code
		if result.StatusCode != 202 { //SUCCESS_CODE
			// TODO: stop execution and inform the user about the errors
			return errors.New("Error starting mappers")
		}
	}

	return c.WriteDoneObject(ctx, "reducers-invoked")
}

// InvokeReducer is used to invoke the final reducer
func (c *Coordinator) InvokeReducer(ctx context.Context, reducerName string) error {
	// check if reducers have been invoked
	if c.GetDoneObject(ctx, "final-reducer-invoked") {
		// reducers have already been invoked
		// it is likely that the coordinator crashed
		return nil
	}

	// function arn
	functionArn := fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s_%s",
		c.Region,
		c.AccountID,
		reducerName,
		c.JobID.String(),
	)

	// encode reducer input to json
	reducerInput := ReducerInput{
		JobID:       c.JobID,
		ReducerID:   uuid.New(),
		NumReducers: int(c.NumQueues),
		NumMappers:  int(c.NumMappers),
	}
	requestPayload, err := json.Marshal(reducerInput)
	if err != nil {
		return err
	}

	result, err := c.FaasAPI.Invoke(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(functionArn),
			Payload:        requestPayload,
			InvocationType: types.InvocationTypeEvent,
		},
	)
	if err != nil {
		return err
	}

	// error is ignored from asynch invokation and result only holds the status code
	// check status code
	if result.StatusCode != 202 { //SUCCESS_CODE
		// TODO: stop execution and inform the user about the errors
		return errors.New("Error starting mappers")
	}

	return c.WriteDoneObject(ctx, "final-reducer-invoked")
}

// AreReducersDone reads events from the reducers-done queue to check
// if all reducers are done
func (c *Coordinator) AreReducersDone(ctx context.Context, nextLogToken *string) (*string, error) {
	queueName := fmt.Sprintf("%s-reducers-done", c.JobID.String())
	queueURL := GetQueueURL(queueName, c.Region, c.AccountID, c.local)
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: MaxItemsPerBatch,
	}

	// keeps a map of done reducers, this is used as the dedupe mechanism
	doneReducers := make(map[string]bool)
	doneReducersCount := 0

	// loop until all reducers are done
	for doneReducersCount < int(c.NumQueues) {
		// reducers are not done yet
		output, err := c.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return nil, err
		}

		if len(output.Messages) != 0 {
			// at least some reducers have completed
			for _, message := range output.Messages {
				// add reducer to done map
				if _, ok := doneReducers[*message.Body]; !ok {
					doneReducers[*message.Body] = true
					doneReducersCount++
				}
			}

			// break and continue processing without sleep
			break
		}

		// sleep for 5 seconds before trying to get more results
		time.Sleep(1 * time.Second)
	}

	messagesReceived := 0
	// loop until all reducers are done
	for doneReducersCount < int(c.NumQueues) {
		// reducers are not done yet
		output, err := c.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, message := range output.Messages {
			// add reducer to done map
			if _, ok := doneReducers[*message.Body]; !ok {
				doneReducers[*message.Body] = true
				doneReducersCount++
			}
		}

		// only log every 100 messages
		if messagesReceived%100 == 0 {
			if reducersCompleted, err := c.GetNumMessagesInQueue(ctx, queueURL); err == nil {
				nextLogToken, _ = c.LogEvent(
					ctx,
					fmt.Sprintf("Reducers completed: %d/%d", reducersCompleted, c.NumQueues),
					nextLogToken,
				)
			}
		}

		messagesReceived = messagesReceived + 10
	}

	return nextLogToken, nil
}

// WriteDoneObject writes a blank object to indicate that the job has invoked
// the mappers, the reducers or indicate the job has finished
func (c *Coordinator) WriteDoneObject(ctx context.Context, filename string) error {
	bucket := c.JobID.String()

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    aws.String(filename),
		Body:   bytes.NewReader([]byte{}),
	}

	_, err := c.UploaderAPI.Upload(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

// GetDoneObject reads a blank object to verify that the job has invoked
// the mappers, the reducers or indicate the job has finished. If the given object
// exists it returns true and false otherwise
func (c *Coordinator) GetDoneObject(ctx context.Context, filename string) bool {
	bucket := c.JobID.String()

	_, err := c.ObjectStoreAPI.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    aws.String(filename),
	})
	if err != nil {
		return false
	}

	return true
}

// // WriteDoneObject writes a blank object to indicate that the job has finished
// func (c *Coordinator) WriteDoneObject2(ctx context.Context) error {
// 	bucket := c.JobID.String()
// 	key := fmt.Sprintf("done-job")

// 	input := &s3.PutObjectInput{
// 		Bucket: &bucket,
// 		Key:    &key,
// 		Body:   bytes.NewReader([]byte{}),
// 	}

// 	_, err := c.UploaderAPI.Upload(ctx, input)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// GetNumMessagesInQueue gets the approximate number of messages in a queue
func (c *Coordinator) GetNumMessagesInQueue(ctx context.Context, queueURL string) (int, error) {
	res, _ := c.QueuesAPI.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: &queueURL,
		AttributeNames: []sqsTypes.QueueAttributeName{
			sqsTypes.QueueAttributeNameApproximateNumberOfMessages,
			sqsTypes.QueueAttributeNameApproximateNumberOfMessagesDelayed,
			sqsTypes.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		},
	})
	totalMappersStr := res.Attributes["ApproximateNumberOfMessages"]
	totalMappersStr = totalMappersStr + res.Attributes["ApproximateNumberOfMessagesDelayed"]
	totalMappersStr = totalMappersStr + res.Attributes["ApproximateNumberOfMessagesNotVisible"]

	return strconv.Atoi(totalMappersStr)
}

// LogEvents logs multiple events to cloudwatch
func (c *Coordinator) LogEvents(ctx context.Context, messages []string, nextSequenceToken *string) (*string, error) {
	// log data
	logGroupName := fmt.Sprintf("%s-log-group", c.JobID.String())
	logStreamName := fmt.Sprintf("%s-log-stream", c.JobID.String())
	currentTimeNano := time.Now().UTC().UnixNano()
	currentTimeMilli := currentTimeNano / 1000000

	// add messages
	events := make([]cloudWatchTypes.InputLogEvent, len(messages))
	for i, message := range messages {
		currentMessage := message
		events[i] = cloudWatchTypes.InputLogEvent{
			Message:   &currentMessage,
			Timestamp: &currentTimeMilli,
		}
	}
	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  &logGroupName,
		LogStreamName: &logStreamName,
	}

	// add next sequence token
	if nextSequenceToken != nil {
		input.SequenceToken = nextSequenceToken
	}

	// send logs
	logRes, err := c.LogsAPI.PutLogEvents(ctx, input)
	if err != nil {
		return nil, err
	}

	return logRes.NextSequenceToken, nil
}

// LogEvents logs a single event to cloudwatch
func (c *Coordinator) LogEvent(ctx context.Context, message string, nextSequenceToken *string) (*string, error) {
	// log data
	logGroupName := fmt.Sprintf("%s-log-group", c.JobID.String())
	logStreamName := fmt.Sprintf("%s-log-stream", c.JobID.String())
	currentTimeNano := time.Now().UnixNano()
	currentTimeMilli := currentTimeNano / 1000000

	input := &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []cloudWatchTypes.InputLogEvent{
			{
				Message:   &message,
				Timestamp: &currentTimeMilli,
			},
		},
		LogGroupName:  &logGroupName,
		LogStreamName: &logStreamName,
	}

	// add next sequence token
	if nextSequenceToken != nil {
		input.SequenceToken = nextSequenceToken
	}

	// send logs
	logRes, err := c.LogsAPI.PutLogEvents(ctx, input)
	if err != nil {
		return nil, err
	}

	return logRes.NextSequenceToken, nil
}

// StartMappers sends the each split into lambda
func (c *Coordinator) StartMappers(ctx context.Context, numQueues int, functionName string) error {
	// check if mappers have been invoked
	if c.GetDoneObject(ctx, "mappers-invoked") {
		// mappers have already been invoked
		// it is likely that the coordinator crashed
		return nil
	}

	mappings, err := c.GetMappings(ctx)
	if err != nil {
		return err
	}

	// function arn
	functionArn := fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		c.Region,
		c.AccountID,
		functionName,
	)

	for _, currentMapping := range mappings {
		// create payload describing split
		input := &MapperInput{
			JobID:     c.JobID,
			Mapping:   *currentMapping,
			NumQueues: int64(numQueues),
		}

		requestPayload, err := json.Marshal(input)
		if err != nil {
			return err
		}

		// send the mapping split into lamda
		_, err = c.FaasAPI.Invoke(
			ctx,
			&lambda.InvokeInput{
				FunctionName:   aws.String(functionArn),
				Payload:        requestPayload,
				InvocationType: types.InvocationTypeEvent,
			},
		)
		if err != nil {
			return err
		}

		// error is ignored from asynch invokation and result only holds the status code
		// check status code
	}

	// write mappers invoked object
	return c.WriteDoneObject(ctx, "mappers-invoked")
}

func (c *Coordinator) GetMappings(ctx context.Context) ([]*Mapping, error) {

	buf := manager.NewWriteAtBuffer([]byte{})
	_, err := c.DownloaderAPI.Download(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(c.JobID.String()),
		Key:    aws.String("mappings"),
	})
	if err != nil {
		return nil, err
	}

	// unmarshal result
	var mappings []*Mapping
	err = json.Unmarshal(buf.Bytes(), &mappings)
	if err != nil {
		return nil, err
	}

	return mappings, nil
}
