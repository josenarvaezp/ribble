package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"gopkg.in/yaml.v3"
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

func ReadConfigFile(ctx context.Context, bucket string, key string, s3Client *s3.Client) (*objectstore.Buckets, error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}
	defer result.Body.Close()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}

	var buckets objectstore.Buckets
	err = yaml.Unmarshal(body, &buckets)
	if err != nil {
		fmt.Println(err)
		// TODO: log error
		return nil, err
	}

	return &buckets, nil
}
