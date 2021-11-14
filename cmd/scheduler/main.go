package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/scheduler"
)

const (
	configFileKey string = "config.yaml"
)

func main() {
	// Test local execution
	jobID := uuid.New()
	scheduler, err := scheduler.NewScheduler(jobID, true) // TODO add client here
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()

	client, err := config.InitLocalClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	configFile, err := config.ReadConfigFile(ctx, "09cd3797-1b53-4c61-b24f-b454bbec73a7", configFileKey, client)
	if err != nil {
		fmt.Println(err)
		return
	}

	objects := objectstore.BucketsToObjects(configFile.Buckets)

	mappings, err := scheduler.GenerateMappings(ctx, objects)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v ", mappings)
}
