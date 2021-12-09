package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/driver"
)

const (
	configBucket       string = "testBucket" // TODO: get as input from user
	configFileKey      string = "config.yaml"
	defaultRegion      string = "eu-west-2"
	mapperFunctionName string = "mapperFuncName"
)

func main() {
	jobID := uuid.New()
	driver, err := driver.NewDriver(jobID, defaultRegion, false /*Not local*/)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	// read config file from bucket
	configFile, err := driver.ReadConfigFile(ctx, configBucket, configFileKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Setting up resources
	err = driver.CreateJobBucket(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = driver.CreateCoordinatorNotification(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = driver.CreateQueues(ctx, 5) // TODO: get num of queues
	if err != nil {
		fmt.Println(err)
		return
	}

	mappings, err := driver.GenerateMappings(ctx, configFile.InputBuckets)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = driver.StartMappers(ctx, mappings, mapperFunctionName)
	if err != nil {
		fmt.Println(err)
	}
}
