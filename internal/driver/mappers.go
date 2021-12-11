package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

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
	MB           int64 = 1048576
	CHUNK_SIZE   int64 = 64 * MB // size of chunks in bytes
	SUCCESS_CODE int32 = 202     // sucessful code for asynchronous lambda invokation
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

// GenerateMappingsCompleteObjects generates map batches such that each individual file is in a single batch.
// Note that if the file doesn't fit in a batch it will be ignored. This allow users to process file where
// the whole file is needed by a single mapper. An example is an aplication where the user wants to process
// images using AI, and for this each image needs to be fed into the algorithm.
func (d *Driver) GenerateMappingsCompleteObjects(ctx context.Context, inputBuckets []*objectstore.Bucket) ([]*mapping, error) {
	// init mappings
	mappings := []*mapping{}
	firstMapping := newMapping()
	lastMapping := firstMapping

	// used for pagination in the list objects call
	var continuationToken *string

	// generate mappings for all buckets
	for i, bucket := range inputBuckets {
		// indifcates if there are more objects to be listed
		moreObjects := true

		for moreObjects {
			params := &s3.ListObjectsV2Input{
				Bucket:            &bucket.Name,
				MaxKeys:           1000,
				ContinuationToken: continuationToken,
			}

			listObjectsOuput, err := d.s3Client.ListObjectsV2(ctx, params)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// update pagination token
			continuationToken = listObjectsOuput.NextContinuationToken

			// check if there are more objects remaining
			moreObjects = listObjectsOuput.IsTruncated

			objects := s3ObjectsToObjects(bucket.Name, listObjectsOuput.Contents)
			partialMappings := generateMappingsForCompleteObjects(objects, lastMapping)

			if !moreObjects && i == len(inputBuckets)-1 {
				// last iteration of list results, add last mapping
				mappings = append(mappings, partialMappings...)
			} else {
				// there are more items for the list operation
				// do not add last mapping as it will be added in the next list
				lastMapping = partialMappings[len(partialMappings)-1]
				partialMappings[len(partialMappings)-1] = nil // Erase element
				partialMappingsMinusLast := partialMappings[:len(partialMappings)-1]

				mappings = append(mappings, partialMappingsMinusLast...)
			}
		}
	}

	return mappings, nil
}

// GenerateMappingsForCompleteObjects is a helper function that generates batches where each file
// needs to fit in a single batch
func generateMappingsForCompleteObjects(objects []objectstore.Object, lastMapping *mapping) []*mapping {
	partialMappings := []*mapping{lastMapping}
	currentMapping := 0

	for _, object := range objects {
		if object.Size > CHUNK_SIZE {
			// object doesn't fit anywhere, ignore object
			// TODO: inform user object doesn't fit
			continue
		}

		availableSpace := CHUNK_SIZE - partialMappings[currentMapping].size
		if object.Size > availableSpace {
			// current object doesn't fit in mapping
			nextMapping := newMapping()
			partialMappings = append(partialMappings, nextMapping)
			currentMapping++
		}

		// add current object to mapping
		objectWithRange := objectstore.NewObjectWithRange(object, 1, object.Size)
		partialMappings[currentMapping].objects = append(partialMappings[currentMapping].objects, objectWithRange)
		partialMappings[currentMapping].size = partialMappings[currentMapping].size + object.Size
	}

	return partialMappings
}

func s3ObjectToObject(bucket string, s3Object s3Types.Object) objectstore.Object {
	return objectstore.Object{
		Bucket: bucket,
		Key:    *s3Object.Key,
		Size:   s3Object.Size,
	}
}

func s3ObjectsToObjects(bucket string, s3Objects []s3Types.Object) []objectstore.Object {
	objects := make([]objectstore.Object, len(s3Objects))
	for i, s3Object := range s3Objects {
		objects[i] = s3ObjectToObject(bucket, s3Object)
	}

	return objects
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
		if result.StatusCode != SUCCESS_CODE {
			// TODO: stop execution and inform the user about the errors
			return errors.New("Error starting mappers")
		}
	}

	return nil
}
