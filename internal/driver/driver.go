package driver

import (
	"context"

	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/faas"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/queues"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

// DriverInterface defines the methods available for the Driver
type DriverInterface interface {
	// build
	BuildJobGenerationBinary() error
	GenerateResourcesFromBinary() error
	BuildDockerImages() error

	// upload
	GenerateMappingsCompleteObjects(ctx context.Context, inputBuckets []*objectstore.Bucket) ([]*Mapping, error)
	CreateJobBucket(ctx context.Context) error
	CreateQueues(ctx context.Context, numQueues int)

	// run
	StartMappers(ctx context.Context, mappings []*Mapping, functionName string, region string) error
}

// Driver is a struct that implements the Driver interface
type Driver struct {
	JobID   uuid.UUID
	JobPath string
	// clients
	ObjectStoreAPI objectstore.ObjectStoreAPI
	FaasAPI        faas.FaasAPI
	QueuesAPI      queues.QueuesAPI
	// user config
	Config Config
}

// NewDriver creates a new Driver struct
func NewDriver(jobID uuid.UUID, conf *Config) (*Driver, error) {
	var cfg aws.Config
	var err error

	// init driver with job id
	driver := &Driver{
		JobID: jobID,
	}

	// set configuration
	driver.Config = *conf

	if driver.Config.Local {
		// point clients to localstack
		cfg, err = config.InitLocalCfg()
		if err != nil {
			return nil, err
		}
	} else {
		// Load the configuration using the aws config file
		cfg, err = config.InitCfg(driver.Config.Region)
		if err != nil {
			return nil, err
		}
	}

	// create and add clients to driver
	driver.ObjectStoreAPI = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	driver.FaasAPI = lambda.NewFromConfig(cfg)
	driver.QueuesAPI = sqs.NewFromConfig(cfg)

	return driver, nil
}
