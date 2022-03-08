package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/josenarvaezp/displ/internal/lambdas"
	"github.com/josenarvaezp/displ/internal/objectstore"
)

// TODOs:
// - When creating the mapper lambda remember to set up the max number of retries and max age of events

const (
	MB           int64 = 1048576
	CHUNK_SIZE   int64 = 64 * MB // size of chunks in bytes
	SUCCESS_CODE int32 = 202     // sucessful code for asynchronous lambda invokation
)

// GenerateMappings generates batches of input data. If logicalSplit is true
// it genererates logical splits (currently only EOL splits are supportes), otherwise
// the mappings are generetated one per file
func (d *Driver) GenerateMappings(ctx context.Context) ([]*lambdas.Mapping, error) {
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
				Bucket:  &bucket,
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

			objects := objectstore.S3ObjectsToObjects(bucket, listObjectsOuput.Contents)

			var partialMappings []*lambdas.Mapping
			var mappingErr error

			if d.Config.LogicalSplit {
				partialMappings, mappingErr = d.generateMappingsForPartialObjects(objects, lastMapping)
				if mappingErr != nil {
					return nil, err
				}
			} else {
				partialMappings = generateMappingsForCompleteObjects(objects, lastMapping)
			}

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

// GenerateMappingsForCompleteObjects is a helper function that generates map batches such that each individual
// file is in a single batch.
// Note that if the file doesn't fit in a batch it will be ignored. This allow users to process file where
// the whole file is needed by a single mapper. An example is an aplication where the user wants to process
// images using AI, and for this each image needs to be fed into the algorithm.
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

// split generates map batches splitting the files in a logical way.
// For example, CSV files split by the end of a line. This splitting allows the framework to process
// massive files (to a maximum of 5 TB) in a distributed way.
func (d *Driver) split(ctx context.Context, object objectstore.Object, initialByte, maxSize int64) (*objectstore.ObjectRange, error) {
	//average min size of a line in a csv file
	lookupRange := 50
	lastByteOfSplit := maxSize

	// usually we should be able to find the last symbol in the last
	// 100 bytes but if not, then double value until we do
	for true {
		// we start looking where we think is big enough to find
		// the last end of line symbol
		lookupRange = lookupRange * 2

		startLookupByte := maxSize + initialByte - int64(lookupRange)

		var buf []byte
		writeAt := manager.NewWriteAtBuffer(buf)

		if startLookupByte > object.Size {
			// add all remaining object
			return &objectstore.ObjectRange{
				Bucket:      object.Bucket,
				Key:         object.Key,
				InitialByte: initialByte,
				FinalByte:   object.Size,
			}, nil
		}

		bytes := fmt.Sprintf("bytes=%d-%d", startLookupByte, maxSize+initialByte)
		if maxSize+initialByte > object.Size {
			// range out of range
			bytes = fmt.Sprintf("bytes=%d-%d", startLookupByte, object.Size)
		}

		_, err := d.DownloaderAPI.Download(ctx, writeAt, &s3.GetObjectInput{
			Bucket: &object.Bucket,
			Key:    &object.Key,
			Range:  aws.String(bytes),
		})
		if err != nil {
			return nil, err
		}

		// get byte of last new line
		indexOfLastEndOfLine := strings.LastIndex(string(writeAt.Bytes()), "\n")
		if indexOfLastEndOfLine == -1 {
			// EOL not found so double the lookup range
			lookupRange = lookupRange * 2
			continue
		}

		// EOL found
		lastByteOfSplit = startLookupByte + int64(indexOfLastEndOfLine)

		break
	}

	return &objectstore.ObjectRange{
		Bucket:      object.Bucket,
		Key:         object.Key,
		InitialByte: initialByte,
		FinalByte:   lastByteOfSplit,
	}, nil
}

// generateMappingsForPartialObjects is a helper function that generates batches where each file
// needs to fit in a single batch
func (d *Driver) generateMappingsForPartialObjects(objects []objectstore.Object, lastMapping *lambdas.Mapping) ([]*lambdas.Mapping, error) {
	partialMappings := []*lambdas.Mapping{lastMapping}
	currentMapping := 0

	for _, object := range objects {
		availableSpace := CHUNK_SIZE - partialMappings[currentMapping].Size

		// split object to fit the current mapping
		splitObjectWithRange, err := d.split(context.Background(), object, 1, availableSpace)
		if err != nil {
			return nil, err
		}

		// add splited object
		partialMappings[currentMapping].Objects = append(partialMappings[currentMapping].Objects, *splitObjectWithRange)
		partialMappings[currentMapping].Size = partialMappings[currentMapping].Size + splitObjectWithRange.FinalByte

		// add the rest of the object
		remainingBytes := object.Size - splitObjectWithRange.FinalByte
		for remainingBytes > 0 {
			// add new mapping
			nextMapping := lambdas.NewMapping()
			partialMappings = append(partialMappings, nextMapping)
			currentMapping++

			splitObjectWithRange, err = d.split(context.Background(), object, splitObjectWithRange.FinalByte, CHUNK_SIZE)
			if err != nil {
				return nil, err
			}

			splitSize := (splitObjectWithRange.FinalByte - splitObjectWithRange.InitialByte)

			partialMappings[currentMapping].Objects = append(partialMappings[currentMapping].Objects, *splitObjectWithRange)
			partialMappings[currentMapping].Size = partialMappings[currentMapping].Size + splitSize

			remainingBytes = remainingBytes - splitSize
		}
	}

	return partialMappings, nil
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
