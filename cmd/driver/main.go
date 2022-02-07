package main

import (
	"context"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/josenarvaezp/displ/internal/driver"
	"github.com/josenarvaezp/displ/internal/logs"
)

func main() {
	jobID := uuid.New()
	ctx := context.Background()

	// get driver config values
	conf, err := driver.ReadLocalConfigFile("config.yaml")
	if err != nil {
		log.WithField(
			"File name", "TODO: get config file name",
		).WithError(err).Error("Error reading config file")
		return
	}

	// set logger
	log.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))
	driverLogger := log.WithFields(log.Fields{
		"Job ID": jobID.String(),
	})

	// init driver
	driver, err := driver.NewDriver(jobID, conf)
	if err != nil {
		driverLogger.WithError(err).Error("Error initializing the driver")
		return
	}

	// Setting up resources
	err = driver.CreateJobBucket(ctx)
	if err != nil {
		driverLogger.WithError(err).Error("Error creating the job bucket")
		return
	}

	// create streams for job
	err = driver.CreateQueues(ctx, 5) // TODO: get num of queues
	if err != nil {
		driverLogger.WithError(err).Error("Error creating the job streams")
		return
	}

	// generate mappings from S3 input bucket
	mappings, err := driver.GenerateMappingsCompleteObjects(ctx)
	if err != nil {
		driverLogger.WithError(err).Error("Error generating mappings from S3")
		return
	}

	err = driver.StartMappers(ctx, mappings, driver.Config.MapperFuncName)
	if err != nil {
		driverLogger.WithError(err).Error("Error starting the mappers")
		return
	}
}
