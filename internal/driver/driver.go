package driver

import (
	"context"
	"fmt"

	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// TODOS:
// add sessions instead of clients sess := session.Must(session.NewSession()) svc = dynamodb.New(sess)

// DriverInterface defines the methods available for the Driver
type DriverInterface interface {
	GenerateMappings(context.Context, []objectstore.Object) ([]mapping, error)
	StartMappers(ctx context.Context, mappings []*mapping, functionName string, region string) error
	CreateJobBucket(ctx context.Context) error
	CreateCoordinatorNotification(ctx context.Context) error
}

// Driver is a struct that implements the Driver interface
type Driver struct {
	jobID        uuid.UUID
	s3Client     *s3.Client
	lambdaClient *lambda.Client
}

// NewDriver creates a new Driver struct
func NewDriver(jobID uuid.UUID, local bool) (Driver, error) {
	var s3Client *s3.Client
	var lambdaClient *lambda.Client
	var err error
	if local {
		s3Client, err = config.InitLocalS3Client()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return Driver{}, err
		}

		lambdaClient, err = config.InitLocalLambdaClient()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return Driver{}, err
		}
	}

	// TODO: implement non local client

	return Driver{
		jobID:        jobID,
		s3Client:     s3Client,
		lambdaClient: lambdaClient,
	}, nil
}
