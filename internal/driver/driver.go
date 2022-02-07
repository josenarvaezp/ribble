package driver

import (
	"context"
	"fmt"

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
	GenerateMappingsCompleteObjects(ctx context.Context, inputBuckets []*objectstore.Bucket) ([]*Mapping, error)
	StartMappers(ctx context.Context, mappings []*Mapping, functionName string, region string) error
	CreateJobBucket(ctx context.Context) error
	CreateCoordinatorNotification(ctx context.Context) error
	CreateQueues(ctx context.Context, numQueues int)
}

// Driver is a struct that implements the Driver interface
type Driver struct {
	jobID uuid.UUID
	// clients
	ObjectStoreAPI objectstore.ObjectStoreAPI
	FaasAPI        faas.FaasAPI
	QueuesAPI      queues.QueuesAPI
	// user config
	Config Config
}

// NewDriver creates a new Driver struct
func NewDriver(jobID uuid.UUID, configFilePath string) (*Driver, error) {
	var cfg aws.Config
	var err error

	// init driver with job id
	driver := &Driver{
		jobID: jobID,
	}

	// set config values to driver
	conf, err := ReadLocalConfigFile(configFilePath)
	if err != nil {
		// TODO: add logs
		fmt.Println(err)
		return nil, err
	}
	driver.Config = *conf

	if driver.Config.Local {
		// point clients to localstack
		cfg, err = config.InitLocalCfg()
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
			return nil, err
		}
	} else {
		// Load the configuration using the aws config file
		cfg, err = config.InitCfg(driver.Config.Region)
		if err != nil {
			// TODO: add logs
			fmt.Println(err)
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
