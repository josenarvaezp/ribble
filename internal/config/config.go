package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitLocalCfg() (aws.Config, error) {
	localstackEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: "https://127.0.0.1:4566",
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

func InitLocalClient() (*s3.Client, error) {
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
