package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/driver"
)

func main() {
	jobID := uuid.New()
	driver, err := driver.NewDriver(jobID, "config.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	//Setting up resources
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

	mappings, err := driver.GenerateMappingsCompleteObjects(ctx, driver.Config.InputBuckets)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = driver.StartMappers(ctx, mappings, driver.Config.MapperFuncName)
	if err != nil {
		fmt.Println(err)
		return
	}
}
