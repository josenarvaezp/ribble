package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/josenarvaezp/displ/internal/lambdas"
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

// GenerateMappingsCompleteObjects generates map batches such that each individual file is in a single batch.
// Note that if the file doesn't fit in a batch it will be ignored. This allow users to process file where
// the whole file is needed by a single mapper. An example is an aplication where the user wants to process
// images using AI, and for this each image needs to be fed into the algorithm.
func (d *Driver) GenerateMappingsCompleteObjects(ctx context.Context) ([]*lambdas.Mapping, error) {
	// init mappings
	mappings := []*lambdas.Mapping{}
	firstMapping := lambdas.NewMapping()
	lastMapping := firstMapping

	// used for pagination in the list objects call
	var continuationToken *string

	// generate mappings for all buckets
	for i, bucket := range d.Config.InputBuckets {
		// indifcates if there are more objects to be listed
		moreObjects := true

		for moreObjects {
			params := &s3.ListObjectsV2Input{
				Bucket:  &bucket.Name,
				MaxKeys: 1000,
			}

			// add continuation token
			if continuationToken != nil {
				params.ContinuationToken = continuationToken
			}

			listObjectsOuput, err := d.ObjectStoreAPI.ListObjectsV2(ctx, params)
			if err != nil {
				return nil, err
			}

			// update pagination token
			continuationToken = listObjectsOuput.NextContinuationToken

			// check if there are more objects remaining
			moreObjects = listObjectsOuput.IsTruncated

			objects := objectstore.S3ObjectsToObjects(bucket.Name, listObjectsOuput.Contents)
			partialMappings := generateMappingsForCompleteObjects(objects, lastMapping)

			if !moreObjects && i == len(d.Config.InputBuckets)-1 {
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
func generateMappingsForCompleteObjects(objects []objectstore.Object, lastMapping *lambdas.Mapping) []*lambdas.Mapping {
	partialMappings := []*lambdas.Mapping{lastMapping}
	currentMapping := 0

	for _, object := range objects {
		if object.Size > CHUNK_SIZE {
			// object doesn't fit anywhere, ignore object
			// TODO: inform user object doesn't fit
			continue
		}

		availableSpace := CHUNK_SIZE - partialMappings[currentMapping].Size
		if object.Size > availableSpace {
			// current object doesn't fit in mapping
			nextMapping := lambdas.NewMapping()
			partialMappings = append(partialMappings, nextMapping)
			currentMapping++
		}

		// add current object to mapping
		objectWithRange := objectstore.NewObjectWithRange(object, 1, object.Size)
		partialMappings[currentMapping].Objects = append(partialMappings[currentMapping].Objects, objectWithRange)
		partialMappings[currentMapping].Size = partialMappings[currentMapping].Size + object.Size
	}

	return partialMappings
}

// StartMappers sends the each split into lambda
func (d *Driver) StartMappers(ctx context.Context, mappings []*lambdas.Mapping, numQueues int) error {
	// function arn
	functionArn := fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s_%s",
		d.Config.Region,
		d.Config.AccountID,
		d.BuildData.MapperData.Function,
		d.JobID.String(),
	)

	for _, currentMapping := range mappings {
		// create payload describing split
		input := &lambdas.MapperInput{
			JobID:     d.JobID,
			Mapping:   *currentMapping,
			NumQueues: int64(numQueues),
		}

		requestPayload, err := json.Marshal(input)
		if err != nil {
			return err
		}

		// send the mapping split into lamda
		_, err = d.FaasAPI.Invoke(
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
		// if result.StatusCode != SUCCESS_CODE {
		// 	// TODO: stop execution and inform the user about the errors
		// 	return errors.New("Error starting mappers")
		// }
	}

	return nil
}
