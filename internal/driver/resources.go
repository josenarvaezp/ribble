package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateJobBucket creates a bucket for the job. This bucket is used as the working directory
// for the job's intermediate output.
func (d *Driver) CreateJobBucket(ctx context.Context) error {
	params := &s3.CreateBucketInput{
		Bucket: &d.jobBucket,
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(d.Config.Region),
		},
	}

	_, err := d.s3Client.CreateBucket(ctx, params, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// CreateCoordinatorNotification adds the configuration to S3 so that when the last mapper
// finish its execution the coordinator is invoked. This is possible because the last
// mapper will create a blank object to indicate it is done. This event is picked by
// the S3 service and invokes the coordinator function.
func (d *Driver) CreateCoordinatorNotification(ctx context.Context) error {
	// TODO: create coordinator IAM role with "s3:GetObject" and resource to the folder under the bucket

	action := "lambda:InvokeFunction"
	coordinatorName := "arn:aws:lambda:eu-west-2:694616335238:function:TODO"
	principal := "s3.amazonaws.com"
	statementId := "s3invoke"
	sourceARN := fmt.Sprintf("arn:aws:s3:::%s", d.jobBucket)

	// add permision to allow S3 to invoke the coordinator function
	// on object creation
	permissionInput := &lambda.AddPermissionInput{
		Action:       &action,
		FunctionName: &coordinatorName,
		Principal:    &principal,
		StatementId:  &statementId,
		SourceArn:    &sourceARN,
	}
	_, err := d.lambdaClient.AddPermission(ctx, permissionInput)
	if err != nil {
		return err
	}

	// location where the last mapper will create the blank object
	prefixForCoordinatorSignal := "signals/coordinator/"

	// add notification configuration so that S3 can invoke the coordinator
	// once an object in signals/coordinator/ has been created. A blank
	// object in this file means that the last mapper has completed execution
	notificationConfigInput := &s3.PutBucketNotificationConfigurationInput{
		Bucket: &d.jobBucket,
		NotificationConfiguration: &s3Types.NotificationConfiguration{
			LambdaFunctionConfigurations: []s3Types.LambdaFunctionConfiguration{
				{
					Events: []s3Types.Event{
						"s3:ObjectCreated:*",
					},
					LambdaFunctionArn: &coordinatorName,
					Filter: &s3Types.NotificationConfigurationFilter{
						Key: &s3Types.S3KeyFilter{
							FilterRules: []s3Types.FilterRule{
								{
									Name:  s3Types.FilterRuleNamePrefix,
									Value: &prefixForCoordinatorSignal,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = d.s3Client.PutBucketNotificationConfiguration(ctx, notificationConfigInput)
	if err != nil {
		fmt.Println(err)
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

	dlqOutput, err := d.sqsClient.CreateQueue(ctx, dlqParams)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// create policy and convert it to json
	dlqARN := GetQueueArnFromURL(dlqOutput.QueueUrl)
	policy := map[string]string{
		"deadLetterTargetArn": *dlqARN,
		"maxReceiveCount":     "3", // three retries
	}

	policyJson, err := json.Marshal(policy)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for i := 1; i <= numQueues; i++ {
		// name of the queues takes the job id as prefix
		currentQueueName := fmt.Sprintf("%s-%d", d.jobID.String(), i)
		params := &sqs.CreateQueueInput{
			QueueName: &currentQueueName,
			Attributes: map[string]string{
				"RedrivePolicy":     string(policyJson),
				"VisibilityTimeout": "60", // TODO: configure
			},
		}
		_, err := d.sqsClient.CreateQueue(ctx, params)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// wait one second before the queues can be used
	time.Sleep(1 * time.Second)

	return nil
}

// GetQueueArnFromURL creates the ARN of a queue based on its URL
func GetQueueArnFromURL(queueURL *string) *string {
	parts := strings.Split(*queueURL, "/")
	subParts := strings.Split(parts[2], ".")

	arn := "arn:aws:" + subParts[0] + ":" + subParts[1] + ":" + parts[3] + ":" + parts[4]

	return &arn
}
