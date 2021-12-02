package scheduler

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

// SchedulerInterface defines the methods available for the scheduler
type SchedulerInterface interface {
	GenerateMappings(context.Context, []objectstore.Object) ([]mapping, error)
	StartMappers(ctx context.Context, mappings []*mapping, functionName string, region string) error
	StartCoordinator(ctx context.Context, functionName string) error
}

// Scheduler is a struct that implements the scheduler interface
type Scheduler struct {
	jobID        uuid.UUID
	s3Client     *s3.Client
	lambdaClient *lambda.Client
}

// NewScheduler creates a new scheduler struct
func NewScheduler(jobID uuid.UUID, local bool) (Scheduler, error) {
	var s3Client *s3.Client
	var lambdaClient *lambda.Client
	var err error
	if local {
		s3Client, err = config.InitLocalS3Client()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return Scheduler{}, err
		}

		lambdaClient, err = config.InitLocalLambdaClient()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return Scheduler{}, err
		}
	}

	// TODO: implement non local client

	return Scheduler{
		jobID:        jobID,
		s3Client:     s3Client,
		lambdaClient: lambdaClient,
	}, nil
}
