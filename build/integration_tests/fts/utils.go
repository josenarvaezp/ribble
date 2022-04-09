package fts

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/stretchr/testify/require"
)

func assertOutput(t *testing.T, expectedOutputFile string, jobID string) {
	cfg, err := config.InitLocalCfg("localhost", 4566, "eu-west-2")
	if err != nil {
		fmt.Println(err)
		return
	}

	// create s3 client
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// wait for job to be completed
	for true {
		objects, err := s3Client.ListObjects(context.Background(), &s3.ListObjectsInput{
			Bucket: &jobID,
			Prefix: aws.String("output/"),
		})
		if err != nil {
			// wait 5 seconds
			time.Sleep(5 * time.Second)
			continue
		}

		// output ready
		require.Len(t, objects.Contents, 1)

		res, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: aws.String(jobID),
			Key:    objects.Contents[0].Key,
		})
		require.Nil(t, err)

		defer res.Body.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		result := buf.String()

		dat, err := os.ReadFile(expectedOutputFile)
		require.Nil(t, err)

		expectedResult := string(dat)

		require.JSONEq(t, expectedResult, result)

		return
	}
}
