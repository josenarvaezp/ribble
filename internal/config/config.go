package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Sources:
// - Setting credentials: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

func InitLocalCfg() (aws.Config, error) {
	localstackEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "https://127.0.0.1:4566",
			SigningRegion: "eu-west-2",
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
		// TODO: log error
		fmt.Println(err)
		return aws.Config{}, err
	}

	// FIXME: insecure client for testing purposes
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	cfg.HTTPClient = httpClient

	return cfg, nil
}

func InitLocalLambdaCfg() (aws.Config, error) {
	localstackEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "local", // TODO
			URL:           "https://localstack:4566",
			SigningRegion: "eu-west-2",
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
		// TODO: log error
		fmt.Println(err)
		return aws.Config{}, err
	}

	// FIXME: insecure client for testing purposes
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	cfg.HTTPClient = httpClient

	return cfg, nil
}

// InitCfg initializes the configuration for the aws services that the driver
// needs. Please note that the AWS credentials are taken from the
// credentials file uner .aws placed in the home directory of the computer
// runnin the driver
func InitCfg(region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		// TODO: log error
		fmt.Println(err)
		return aws.Config{}, err
	}

	return cfg, nil
}

func InitLocalS3Client() (*s3.Client, error) {
	cfg, err := InitLocalCfg()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

func InitLocalLambdaClient() (*lambda.Client, error) {
	cfg, err := InitLocalCfg()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return lambda.NewFromConfig(cfg), nil
}
