package objectstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ObjectStoreAPI is an interface used to mock API calls made to the aws S3 service
type ObjectStoreAPI interface {
	ListObjectsV2(
		ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options),
	) (*s3.ListObjectsV2Output, error)
	CreateBucket(
		ctx context.Context,
		params *s3.CreateBucketInput,
		optFns ...func(*s3.Options),
	) (*s3.CreateBucketOutput, error)
}

// Object represent a cloud object
type Object struct {
	Bucket string
	Key    string
	Size   int64
}

// ObjectRange represents an cloud object with its range specified
type ObjectRange struct {
	Bucket      string `json:"objectBucket"`
	Key         string `json:"objectKey"`
	InitialByte int64  `json:"initialByte"`
	FinalByte   int64  `json:"finalByte"`
}

// Bucket represents a cloud bucket
type Bucket struct {
	Name string `yaml:"bucket"`
	Keys []string
}

// NewObjectWithRange creates a new objectRange
func NewObjectWithRange(object Object, initialByte int64, finalByte int64) ObjectRange {
	return ObjectRange{
		Bucket:      object.Bucket,
		Key:         object.Key,
		InitialByte: initialByte,
		FinalByte:   finalByte,
	}
}

func s3ObjectToObject(bucket string, s3Object types.Object) Object {
	return Object{
		Bucket: bucket,
		Key:    *s3Object.Key,
		Size:   s3Object.Size,
	}
}

func S3ObjectsToObjects(bucket string, s3Objects []types.Object) []Object {
	objects := make([]Object, len(s3Objects))
	for i, s3Object := range s3Objects {
		objects[i] = s3ObjectToObject(bucket, s3Object)
	}

	return objects
}
