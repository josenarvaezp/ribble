package scheduler

import (
	"context"
	"fmt"

	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// SchedulerInterface defines the methods available for the scheduler
type SchedulerInterface interface {
	GenerateMappings(context.Context, []objectstore.Object) ([]mapping, error)
}

// Scheduler is a struct that implements the scheduler interface
type Scheduler struct {
	jobID    uuid.UUID
	s3Client *s3.Client
}

// NewScheduler creates a new scheduler struct
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
