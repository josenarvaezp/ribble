package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
// tutorial on creating an event map source for coordinator: https://docs.aws.amazon.com/lambda/latest/dg/with-s3-tutorial.html

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
func (s *Driver) GenerateMappings(ctx context.Context, inputBuckets *objectstore.Buckets) ([]*mapping, error) {
	// init mappings
	currentMapping := 0
	mappings := []*mapping{}
	firstMapping := newMapping()
	mappings = append(mappings, firstMapping)

	// get all objects across buckets
	objects := objectstore.BucketsToObjects(inputBuckets.Buckets)

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
		if result.StatusCode == successCode {
			// TODO: update task to "Received" status in dynamoDB
		} else {
			// TODO: stop execution and inform the user about the errors
		}
	}

	return nil
}

// CreateJobBucket creates a bucket for the job. This bucket is used as the working directory
// for the job's intermediate output.
func (d *Driver) CreateJobBucket(ctx context.Context) error {
	bucketName := fmt.Sprintf("%s-%s", "displ", d.jobID.String())
	params := &s3.CreateBucketInput{
		Bucket: &bucketName,
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			// TODO: add region programatically
			LocationConstraint: s3Types.BucketLocationConstraintEuWest2,
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
	bucketName := "TODO"

	action := "lambda:InvokeFunction"
	coordinatorName := "arn:aws:lambda:eu-west-2:694616335238:function:TODO"
	principal := "s3.amazonaws.com"
	statementId := "s3invoke"
	sourceARN := fmt.Sprintf("arn:aws:s3:::%s", bucketName)

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
		Bucket: &bucketName,
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
	for i := 1; i <= numQueues; i++ {
		// name of the queues takes the job id as prefix
		currentQueueName := fmt.Sprintf("%s-%d", d.jobID, i)
		params := &sqs.CreateQueueInput{
			QueueName: &currentQueueName,
			// TODO: add attributes and tags
		}
		_, err := d.sqsClient.CreateQueue(ctx, params)
		if err != nil {
			// TODO: maybe retry on error?
			fmt.Println(err)
			return err
		}
	}

	// wait one second before the queues can be used
	time.Sleep(1 * time.Second)

	return nil
}

type ConfigFile struct {
	InputBuckets *objectstore.Buckets
}

// ReadConfigFile reads the config file from the specified object
func (d *Driver) ReadConfigFile(ctx context.Context, bucket string, key string) (*ConfigFile, error) {
	// TODO: extend function to read more config values
	result, err := d.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}
	defer result.Body.Close()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}

	var buckets objectstore.Buckets
	err = yaml.Unmarshal(body, &buckets)
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}

	return &ConfigFile{
		InputBuckets: &buckets,
	}, nil
}
