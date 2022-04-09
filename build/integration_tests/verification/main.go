package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func main() {

	fmt.Println(uuid.New().String())
	fmt.Println(uuid.New().String())
	// parse job id
	var jobID string
	var testID string
	flag.StringVar(&jobID, "job-id", "", "The ID for the job")
	flag.StringVar(&testID, "test-id", "", "The ID for the job")
	flag.Parse()

	cfg, err := InitLocalCfg("localhost", 4566, "eu-west-2")
	if err != nil {
		fmt.Println(err)
		return
	}

	// create s3 client
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	objects, err := s3Client.ListObjects(context.Background(), &s3.ListObjectsInput{
		Bucket: &jobID,
		Prefix: aws.String("output/"),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(objects.Contents) != 1 {
		fmt.Println("error not expected objects!!!")
		return
	}

	res, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(jobID),
		Key:    objects.Contents[0].Key,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	result := buf.String()

	dat, err := os.ReadFile("./build/integration_tests/test_output/test1_out")
	if err != nil {
		fmt.Println(err)
		return
	}

	expectedResult := string(dat)

	fmt.Println(expectedResult)

	// err := json.Unmarshal(result)

	if result != expectedResult {
		fmt.Println("error!!!")
	}

}

func InitLocalCfg(hostEndpoint string, hostPort int, region string) (*aws.Config, error) {
	localstackEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           fmt.Sprintf("https://%s:%d", hostEndpoint, hostPort),
			SigningRegion: region,
		}, nil
	})

	localstackCredentialsResolver := aws.CredentialsProviderFunc(func(context context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     "dummyKey",
			SecretAccessKey: "dummyKey",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolver(localstackEndpointResolver),
		config.WithCredentialsProvider(localstackCredentialsResolver),
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// FIXME: insecure client for testing purposes
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	cfg.HTTPClient = httpClient

	return &cfg, nil
}
