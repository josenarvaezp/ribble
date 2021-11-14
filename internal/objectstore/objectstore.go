package objectstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Object represent a cloud object
type Object struct {
	Bucket string
	Key    string
}

// ObjectRange represents an cloud object with its range specified
type ObjectRange struct {
	Object      Object
	InitialByte int64
	FinalByte   int64
}

// Buckets represents a collection of cloud buckets
type Buckets struct {
	Buckets []Bucket `yaml:"buckets"`
}

// Bucket represents a cloud bucket
type Bucket struct {
	Name string   `yaml:"name"`
	Keys []string `yaml:"keys"`
}

func BucketsToObjects(buckets []Bucket) []Object {
	objects := []Object{}

	for _, bucket := range buckets {
		object := BucketToObject(bucket)
		objects = append(objects, object...)
	}

	return objects
}

func BucketToObject(bucket Bucket) []Object {
	objects := make([]Object, len(bucket.Keys))

	for i, key := range bucket.Keys {
		newObject := Object{
			Bucket: bucket.Name,
			Key:    key,
		}
		objects[i] = newObject
	}

	return objects
}

// GetObjectSize returns the size of the s3 object specified in bytes
func GetObjectSize(ctx context.Context, client *s3.Client, object Object) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: &object.Bucket,
		Key:    &object.Key,
	}
	output, err := client.HeadObject(ctx, input)
	if err != nil {
		// TODO: add logging
		fmt.Println(err)
		return 0, err
	}

	return output.ContentLength, nil
}

// NewObjectWithRange creates a new objectRange
func NewObjectWithRange(object Object, initialByte int64, finalByte int64) ObjectRange {
	return ObjectRange{
		Object:      object,
		InitialByte: initialByte,
		FinalByte:   finalByte,
	}
}
