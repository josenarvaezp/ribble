package fts

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/driver"
	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/stretchr/testify/require"
)

func TestBuildQ6(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	dir, _ := os.Getwd()
	fmt.Println(dir)

	jobPath := "./build/integration_tests/ribble_jobs/query6/job/query6_job.go"

	// set driver
	jobID := uuid.MustParse("2f18f4e1-60e7-491f-97b7-989c996f577e")
	jobDriver := driver.NewBuildDriver(jobID)

	// add job path info to driver
	jobDriver.BuildData = &generators.BuildData{
		JobPath:  jobPath,
		BuildDir: fmt.Sprintf("./build/lambda_gen/%s", jobDriver.JobID.String()),
	}

	// build directory for job's generated files
	if _, err := os.Stat(jobDriver.BuildData.BuildDir); os.IsNotExist(err) {
		err := os.MkdirAll(jobDriver.BuildData.BuildDir, os.ModePerm)
		require.Nil(t, err)
	}

	// build binary that creates lambda files
	err := jobDriver.BuildJobGenerationBinary()
	require.Nil(t, err)

	// run binary to create lambda files (go files and dockerfiles)
	fmt.Println("Generating resources...")
	err = jobDriver.GenerateResourcesFromBinary()
	require.Nil(t, err)

	// build mapper and coordinator docker images
	fmt.Println("Building docker images...")
	err = jobDriver.BuildDockerCustomImages()
	require.Nil(t, err)
}

func TestUploadQ6(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	jobID := uuid.MustParse("2f18f4e1-60e7-491f-97b7-989c996f577e")
	ctx := context.Background()

	// get driver config values
	configFile := fmt.Sprintf("%s/%s/config.yaml", generators.GeneratedFilesDir, jobID)
	conf, err := config.ReadLocalConfigFile(configFile)
	require.Nil(t, err)

	// set driver
	jobDriver, err := driver.NewDriver(jobID, conf)
	require.Nil(t, err)
	jobDriver.JobID = jobID

	// get build data
	buildData, err := generators.ReadBuildData(jobDriver.JobID.String())
	require.Nil(t, err)
	jobDriver.BuildData = buildData

	// create log group and stream
	err = jobDriver.CreateLogsInfra(ctx)
	require.Nil(t, err)

	// Setting up resources
	err = jobDriver.CreateJobBucket(ctx)
	require.Nil(t, err)

	// create dlq SQS for the mappers and coordinator
	dlqArn, err := jobDriver.CreateLambdaDLQ(ctx)
	require.Nil(t, err)

	// upload images to amazon ECR and create lambda function
	err = jobDriver.UploadLambdaFunctions(ctx, dlqArn)
	require.Nil(t, err)
}

func TestRunQ6(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	jobID := uuid.MustParse("2f18f4e1-60e7-491f-97b7-989c996f577e")
	ctx := context.Background()
	reducers := 0

	// get driver config values
	configFile := fmt.Sprintf("%s/%s/config.yaml", generators.GeneratedFilesDir, jobID)
	conf, err := config.ReadLocalConfigFile(configFile)
	require.Nil(t, err)

	// set driver
	jobDriver, err := driver.NewDriver(jobID, conf)
	require.Nil(t, err)
	jobDriver.JobID = jobID

	// get build data
	buildData, err := generators.ReadBuildData(jobDriver.JobID.String())
	require.Nil(t, err)
	jobDriver.BuildData = buildData

	// generate mappings from S3 input bucket
	mappings, err := jobDriver.GenerateMappings(ctx)
	require.Nil(t, err)

	numMappings := len(mappings)
	if reducers == 0 {
		// no reducers specified
		reducers = int(math.Ceil(float64(numMappings) / 2))
	}

	totalObjects := 0
	for _, mapping := range mappings {
		totalObjects = totalObjects + len(mapping.Objects)
	}

	// create streams for job
	err = jobDriver.CreateQueues(ctx, reducers)
	require.Nil(t, err)

	// start coordinator
	err = jobDriver.StartCoordinator(ctx, numMappings, reducers)
	require.Nil(t, err)

	// wait until job has completed
	assertOutputQ6(t, "./build/integration_tests/tests_expected_output/test6_out", jobID.String())
}
