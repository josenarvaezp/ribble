package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/josenarvaezp/displ/internal/lambdas"
)

// CreateJobBucket creates a bucket for the job. This bucket is used as the working directory
// for the job's intermediate output.
func (d *Driver) CreateJobBucket(ctx context.Context) error {
	bucket := d.JobID.String()
	params := &s3.CreateBucketInput{
		Bucket: &bucket,
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(d.Config.Region),
		},
	}

	_, err := d.ObjectStoreAPI.CreateBucket(ctx, params)
	if err != nil {
		// only ingore already exists bucket error
		if !bucketAlreadyExists(err) {
			return err
		}
	}

	return nil
}

// CreateDQL creates the dead-letter queue for the service
func (d *Driver) CreateAggregatorsDLQ(ctx context.Context) (*string, error) {
	// create final reduce queue
	finalMetadataQueueName := fmt.Sprintf("%s-final-aggregator-meta", d.JobID.String())
	_, err := d.QueuesAPI.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: &finalMetadataQueueName,
	})
	if err != nil {
		return nil, err
	}

	finalQueueName := fmt.Sprintf("%s-%s", d.JobID.String(), "final-aggregator")
	_, err = d.QueuesAPI.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: &finalQueueName,
	})
	if err != nil {
		return nil, err
	}

	// create dead-letter queue
	dlqName := "ribble_aggregators_dlq"
	dlqParams := &sqs.CreateQueueInput{
		QueueName: &dlqName,
	}

	dlqOutput, err := d.QueuesAPI.CreateQueue(ctx, dlqParams)
	if err != nil {
		return nil, err
	}

	// create policy and convert it to json
	getQueueAttributesParams := &sqs.GetQueueAttributesInput{
		QueueUrl:       dlqOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	}
	attributes, err := d.QueuesAPI.GetQueueAttributes(ctx, getQueueAttributesParams)
	if err != nil {
		return nil, err
	}

	dlqARN := attributes.Attributes["QueueArn"]
	return &dlqARN, err
}

// CreateDQL creates the dead-letter queue for the service
func (d *Driver) CreateLambdaDLQ(ctx context.Context) (*string, error) {
	// create dead-letter queue
	dlqName := fmt.Sprintf("%s-%s", d.JobID.String(), "lambda-dlq")
	dlqParams := &sqs.CreateQueueInput{
		QueueName: &dlqName,
	}

	dlqOutput, err := d.QueuesAPI.CreateQueue(ctx, dlqParams)
	if err != nil {
		return nil, err
	}

	// create policy and convert it to json
	getQueueAttributesParams := &sqs.GetQueueAttributesInput{
		QueueUrl:       dlqOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	}
	attributes, err := d.QueuesAPI.GetQueueAttributes(ctx, getQueueAttributesParams)
	if err != nil {
		return nil, err
	}

	dlqARN := attributes.Attributes["QueueArn"]
	return &dlqARN, err
}

// CreateQueues creates numQueues. This queues will be used by the framework
// to send data from the mappers to the reducers.
func (d *Driver) CreateQueues(ctx context.Context, numQueues int) error {
	// create dead-letter queue
	dlqName := fmt.Sprintf("%s-%s", d.JobID.String(), "messages-dlq")
	dlqParams := &sqs.CreateQueueInput{
		QueueName: &dlqName,
	}

	dlqOutput, err := d.QueuesAPI.CreateQueue(ctx, dlqParams)
	if err != nil {
		return err
	}

	// create policy and convert it to json
	getQueueAttributesParams := &sqs.GetQueueAttributesInput{
		QueueUrl:       dlqOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	}
	attributes, err := d.QueuesAPI.GetQueueAttributes(ctx, getQueueAttributesParams)
	if err != nil {
		return err
	}
	dlqARN := attributes.Attributes["QueueArn"]

	policy := map[string]string{
		"deadLetterTargetArn": dlqARN,
		"maxReceiveCount":     "3", // three retries
	}

	policyJson, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	// Create a queue used by mappers to indicate they have completed processing
	mappersDoneName := fmt.Sprintf("%s-mappers-done", d.JobID.String())
	mappersDoneParams := &sqs.CreateQueueInput{
		QueueName: &mappersDoneName,
	}

	_, err = d.QueuesAPI.CreateQueue(ctx, mappersDoneParams)
	if err != nil {
		return err
	}

	// Create a queue used by reducers to indicate they have completed processing
	reducersDoneName := fmt.Sprintf("%s-reducers-done", d.JobID.String())
	reducerDoneParams := &sqs.CreateQueueInput{
		QueueName: &reducersDoneName,
	}

	_, err = d.QueuesAPI.CreateQueue(ctx, reducerDoneParams)
	if err != nil {
		return err
	}

	for i := 0; i < numQueues; i++ {
		// create queues where data from mappers will be sent to
		// name of the queues takes the job id as prefix
		currentQueueName := fmt.Sprintf("%s-%d", d.JobID.String(), i)
		params := &sqs.CreateQueueInput{
			QueueName: &currentQueueName,
			Attributes: map[string]string{
				"RedrivePolicy":     string(policyJson),
				"VisibilityTimeout": "60", // TODO: configure
			},
		}
		_, err := d.QueuesAPI.CreateQueue(ctx, params)
		if err != nil {
			return err
		}

		// create a metadata queue for each queue
		currentMetadataQueueName := fmt.Sprintf("%s-%d-meta", d.JobID.String(), i)
		metaParams := &sqs.CreateQueueInput{
			QueueName: &currentMetadataQueueName,
		}
		_, err = d.QueuesAPI.CreateQueue(ctx, metaParams)
		if err != nil {
			return err
		}
	}

	// wait one second before the queues can be used
	time.Sleep(1 * time.Second)

	return nil
}

// StartCoordinator starts a job coordinator
func (d *Driver) StartCoordinator(ctx context.Context, numMappers int, numQueues int) error {
	// coordinator input
	request := &lambdas.CoordinatorInput{
		JobID:      d.JobID,
		NumMappers: numMappers,
		NumQueues:  numQueues,
	}

	// create payload
	requestPayload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// function arn
	functionArn := fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s_%s",
		d.Config.Region,
		d.Config.AccountID,
		d.BuildData.CoordinatorData.Function,
		d.JobID.String(),
	)

	// send the mapping split into lamda
	_, err = d.FaasAPI.Invoke(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(functionArn),
			Payload:        requestPayload,
			InvocationType: lambdaTypes.InvocationTypeEvent,
		},
	)
	return err

	// error is ignored from asynch invokation and result only holds the status code
	// check status code
	// if result.StatusCode != SUCCESS_CODE {
	// 	return errors.New("Error starting coordintator")
	// }
}

// bucketAlreadyExists checks if the s3 bucket being created already exists
func bucketAlreadyExists(err error) bool {
	var alreadyExists *s3Types.BucketAlreadyExists
	return errors.As(err, &alreadyExists)
}
