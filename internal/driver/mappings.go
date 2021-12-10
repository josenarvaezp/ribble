package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"github.com/josenarvaezp/displ/internal/objectstore"
)

// TODOs:
// - This operation requires permission for the lambda:InvokeFunction action.
// - When creating the mapper lambda remember to set up the max number of retries and max age of events
// - reserve concurrency for the coordinator (just one)

// 1. Create a IAM for the coordinator function. The function needs to have access to
// Get the object from the source S3 bucket.
// "s3:GetObject"
// "Resource": "arn:aws:s3:::mybucket/*"

// Sources:
// - tutorial on creating an event map source for coordinator: https://docs.aws.amazon.com/lambda/latest/dg/with-s3-tutorial.html
// -DLQ: https://aws.github.io/aws-sdk-go-v2/docs/code-examples/sqs/deadletterqueue/

const (
	MB          int64 = 1048576
	chunkSize   int64 = 64 * MB // size of chunks in bytes
	successCode int32 = 202     // sucessful code for asynchronous lambda invokation
)

// mapping represents the collection of objects that are used as input
// for the mapping stage of the framework. Each mapper recieves an
// input which may contain one or multiple objects, depeding on their size.
type mapping struct {
	mapID   uuid.UUID
	objects []objectstore.ObjectRange
	size    int64
}

// newMapping initialises the mapping with an id and size 0
func newMapping() *mapping {
	return &mapping{
		mapID: uuid.New(),
		size:  0,
	}
}

// generateMapping generates the input for the mappers. Each mapping has a map id, a list of objects
// where each object has a specified range and the size of it.
func (s *Driver) GenerateMappings(ctx context.Context, inputBuckets []*objectstore.Bucket) ([]*mapping, error) {
	// init mappings
	currentMapping := 0
	mappings := []*mapping{}
	firstMapping := newMapping()
	mappings = append(mappings, firstMapping)

	// get all objects across buckets
	objects := objectstore.BucketsToObjects(inputBuckets)

	for _, object := range objects {
		availableSpace := chunkSize - mappings[currentMapping].size
		// move to next mapping if there is no space in current mapping
		if availableSpace == 0 {
			nextMapping := newMapping()
			mappings = append(mappings, nextMapping)
			currentMapping++
		}

		// read object size from s3
		currentObjectSize, err := objectstore.GetObjectSize(ctx, s.s3Client, object)
		if err != nil {
			fmt.Println(err)
			// TODO: info should be logged if we get this error
			return nil, err
		}

		if currentObjectSize <= availableSpace {
			// current object fits in mapping
			objectWithRange := objectstore.NewObjectWithRange(object, 1, currentObjectSize)
			mappings[currentMapping].objects = append(mappings[currentMapping].objects, objectWithRange)
			mappings[currentMapping].size = mappings[currentMapping].size + currentObjectSize
		} else {
			// current object doesn't fit in mapping
			var mappedBytes int64
			mappedBytes = 0
			remainingBytes := currentObjectSize

			// loop until there are no bytes to map for the current object
			for remainingBytes != 0 {
				if remainingBytes <= availableSpace {
					// remaining bytes fit in the current mapping
					objectWithRange := objectstore.NewObjectWithRange(object, mappedBytes+1, remainingBytes)
					mappings[currentMapping].objects = append(mappings[currentMapping].objects, objectWithRange)
					mappings[currentMapping].size = mappings[currentMapping].size + remainingBytes
					break
				} else {
					// remaining bytes don't fit in current mapping, add as much as possible
					objectWithRange := objectstore.NewObjectWithRange(object, mappedBytes+1, mappedBytes+availableSpace)
					mappings[currentMapping].objects = append(mappings[currentMapping].objects, objectWithRange)
					mappings[currentMapping].size = mappings[currentMapping].size + availableSpace
					mappedBytes = mappedBytes + availableSpace
					remainingBytes = remainingBytes - availableSpace

					// move to next mapping
					nextMapping := newMapping()
					mappings = append(mappings, nextMapping)
					currentMapping++

					availableSpace = chunkSize - mappings[currentMapping].size
				}
			}
		}
	}
	return mappings, nil
}

// StartMappers sends the each split into lambda
func (s *Driver) StartMappers(ctx context.Context, mappings []*mapping, functionName string) error {
	for _, currentMapping := range mappings {
		// create payload describing split
		requestPayload, err := json.Marshal(*currentMapping)
		if err != nil {
			fmt.Println("Error marshalling mapping")
			return err
		}

		// send the mapping split into lamda
		result, _ := s.lambdaClient.Invoke(
			ctx,
			&lambda.InvokeInput{
				FunctionName:   aws.String(functionName),
				Payload:        requestPayload,
				InvocationType: types.InvocationTypeEvent,
			},
		)

		// error is ignored from asynch invokation and result only holds the status code
		// check status code
		if result.StatusCode != successCode {
			// TODO: stop execution and inform the user about the errors
			return errors.New("Error starting mappers")
		}
	}

	return nil
}

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

// GetQueueArnFromURL creates the ARN of a queue based on its URL
func GetQueueArnFromURL(queueURL *string) *string {
	parts := strings.Split(*queueURL, "/")
	subParts := strings.Split(parts[2], ".")

	arn := "arn:aws:" + subParts[0] + ":" + subParts[1] + ":" + parts[3] + ":" + parts[4]

	return &arn
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

type Config struct {
	InputBuckets   []*objectstore.Bucket `yaml:"input"`
	OutputBucket   string                `yaml:"output"`
	Region         string                `yaml:"region"`
	MapperFuncName string                `yaml:"mapperFuncName"`
	Local          bool                  `yaml:"local"`
}

// ReadLocalConfigFile reads the config file from the driver's file system
// note that the path can be absolute or relative path
func ReadLocalConfigFile(path string) (*Config, error) {
	var conf Config

	confFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = yaml.Unmarshal(confFile, &conf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &conf, nil
}
