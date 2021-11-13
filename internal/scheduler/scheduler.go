package scheduler

import (
	"context"
	"fmt"

	"github.com/josenarvaezp/displ/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	MB        int64 = 1048576
	chunkSize int64 = 64 * MB // size of chunks in bytes
)

type Scheduler struct {
	jobID    uuid.UUID
	s3Client *s3.Client
}

func NewScheduler(jobID uuid.UUID, local bool) (Scheduler, error) {
	var client *s3.Client
	var err error
	if local {
		client, err = config.InitLocalClient()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return Scheduler{}, err
		}
	}

	// TODO: implement non local client

	return Scheduler{
		jobID:    jobID,
		s3Client: client,
	}, nil
}

// mapping represents the collection of objects that are used as input
// for the mapping stage of the framework. Each mapper recieves an
// input which may contain one or multiple objects, depeding on their size.
type mapping struct {
	mapID   uuid.UUID
	objects []objectRange
	size    int64
}

// s3Object represent an aws s3 object
type s3Object struct {
	bucket string
	key    string
}

// objectRange represents an s3object with its range specified
type objectRange struct {
	object      s3Object
	initialByte int64
	finalByte   int64
}

// generateMapping generates the input for the mappers. Each mapping has a map id, a list of objects
// where each object has a specified range and the size of it.
func (s *Scheduler) GenerateMappings(ctx context.Context, objects []s3Object) ([]mapping, error) {
	// init mappings
	currentMapping := 0
	mappings := []mapping{}
	firstMapping := initMapping()
	mappings = append(mappings, firstMapping)

	for _, object := range objects {
		availableSpace := chunkSize - mappings[currentMapping].size
		// move to next mapping if there is no space in current mapping
		if availableSpace == 0 {
			nextMapping := initMapping()
			mappings = append(mappings, nextMapping)
			currentMapping++
		}

		// read object size from s3
		currentObjectSize, err := s.getObjectSize(ctx, object)
		if err != nil {
			fmt.Println(err)
			// TODO: info should be logged if we get this error
			return nil, err
		}

		if currentObjectSize <= availableSpace {
			// current object fits in mapping
			objectWithRange := newObjectWithRange(object, 1, currentObjectSize)
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
					objectWithRange := newObjectWithRange(object, mappedBytes+1, remainingBytes)
					mappings[currentMapping].objects = append(mappings[currentMapping].objects, objectWithRange)
					mappings[currentMapping].size = mappings[currentMapping].size + remainingBytes
					break
				} else {
					// remaining bytes don't fit in current mapping, add as much as possible
					objectWithRange := newObjectWithRange(object, mappedBytes+1, mappedBytes+availableSpace)
					mappings[currentMapping].objects = append(mappings[currentMapping].objects, objectWithRange)
					mappings[currentMapping].size = mappings[currentMapping].size + availableSpace
					mappedBytes = mappedBytes + availableSpace
					remainingBytes = remainingBytes - availableSpace

					// move to next mapping
					nextMapping := initMapping()
					mappings = append(mappings, nextMapping)
					currentMapping++

					availableSpace = chunkSize - mappings[currentMapping].size
				}
			}
		}
	}
	return mappings, nil
}

// getObjectSize returns the size of the s3 object specified in bytes
func (s *Scheduler) getObjectSize(ctx context.Context, object s3Object) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: &object.bucket,
		Key:    &object.key,
	}
	output, err := s.s3Client.HeadObject(ctx, input)
	if err != nil {
		// TODO: add logging
		fmt.Println(err)
		return 0, err
	}

	return output.ContentLength, nil
}

// newObjectWithRange creates a new objectRange
func newObjectWithRange(object s3Object, initialByte int64, finalByte int64) objectRange {
	return objectRange{
		object:      object,
		initialByte: initialByte,
		finalByte:   finalByte,
	}
}

// initMapping initialises the mapping with an id and size 0
func initMapping() mapping {
	return mapping{
		mapID: uuid.New(),
		size:  0,
	}
}

// TODO: remove method, only for local testing
func (s *Scheduler) GenerateMockObjects(bucket string, keys ...string) []s3Object {
	objects := make([]s3Object, len(keys))
	for i, key := range keys {
		objects[i] = s3Object{
			bucket: bucket,
			key:    key,
		}
	}

	return objects
}
