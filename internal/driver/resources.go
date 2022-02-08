package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// CreateJobBucket creates a bucket for the job. This bucket is used as the working directory
// for the job's intermediate output.
func (d *Driver) CreateJobBucket(ctx context.Context) error {
	bucket := d.jobID.String()
	params := &s3.CreateBucketInput{
		Bucket: &bucket,
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(d.Config.Region),
		},
	}

	_, err := d.ObjectStoreAPI.CreateBucket(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// CreateQueues creates numQueues. This queues will be used by the framework
// to send data from the mappers to the reducers.
func (d *Driver) CreateQueues(ctx context.Context, numQueues int) error {
	// create dead-letter queue
	dlqName := fmt.Sprintf("%s-%s", d.jobID.String(), "dlq")
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
	mappersDoneName := fmt.Sprintf("%s-mappers-done", d.jobID.String())
	mappersDoneParams := &sqs.CreateQueueInput{
		QueueName: &mappersDoneName,
	}

	_, err = d.QueuesAPI.CreateQueue(ctx, mappersDoneParams)
	if err != nil {
		return err
	}

	// Create a queue used by reducers to indicate they have completed processing
	reducersDoneName := fmt.Sprintf("%s-reducers-done", d.jobID.String())
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
		currentQueueName := fmt.Sprintf("%s-%d", d.jobID.String(), i)
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
		currentMetadataQueueName := fmt.Sprintf("%s-%d-meta", d.jobID.String(), i)
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
