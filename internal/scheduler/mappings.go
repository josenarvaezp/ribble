package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/google/uuid"

	"github.com/josenarvaezp/displ/internal/objectstore"
)

// TODOs:
// - This operation requires permission for the lambda:InvokeFunction action.
// - When creating the mapper lambda remember to set up the max number of retries and max age of events
// - reserve concurrency for the coordinator (just one)

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
func (s *Scheduler) GenerateMappings(ctx context.Context, objects []objectstore.Object) ([]*mapping, error) {
	// init mappings
	currentMapping := 0
	mappings := []*mapping{}
	firstMapping := newMapping()
	mappings = append(mappings, firstMapping)

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
func (s *Scheduler) StartMappers(ctx context.Context, mappings []*mapping, functionName string) error {
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
			// TODO: don't stop execution but inform the user about the errors
		}
	}

	return nil
}

// StartCoordinator starts the lambda coordinator
func (s *Scheduler) StartCoordinator(ctx context.Context, functionName string) error {
	// add job id to coordinator payload
	requestPayload, err := json.Marshal(map[string]string{"job-id": s.jobID.String()})
	if err != nil {
		fmt.Println("Error marshalling job-id")
		return err
	}

	// invoke coordinator lambda
	result, _ := s.lambdaClient.Invoke(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(functionName),
			Payload:        requestPayload,
			InvocationType: types.InvocationTypeEvent,
		},
	)

	if result.StatusCode == successCode {
		// TODO: do something
	} else {
		// we need to invoke the coordinator again
	}

	return nil
}
