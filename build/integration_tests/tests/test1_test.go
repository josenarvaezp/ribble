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

func TestBuildQ1(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	jobPath := "./build/integration_tests/ribble_jobs/query1/job/query1_job.go"

	// set driver
	jobID := uuid.MustParse("88cc574a-83b1-40fa-92fc-3b4d4fd24624")
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

func TestUploadQ1(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	jobID := uuid.MustParse("88cc574a-83b1-40fa-92fc-3b4d4fd24624")
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

	// Setting up resources
	err = jobDriver.CreateJobBucket(ctx)
	require.Nil(t, err)

	// generate mappings from S3 input bucket
	mappings, err := jobDriver.GenerateMappings(ctx)
	require.Nil(t, err)

	// get number of reducers
	numMappings := len(mappings)

	// no reducers specified
	reducers := int(math.Ceil(float64(numMappings) / 2))

	// update build data
	buildData.NumMappers = numMappings
	buildData.NumReducers = reducers
	err = generators.WriteBuildData(buildData, jobID.String())
	require.Nil(t, err)

	// write mappings to s3
	err = jobDriver.WriteMappings(ctx, mappings)
	require.Nil(t, err)

	// create streams for job
	err = jobDriver.CreateQueues(ctx, reducers)
	require.Nil(t, err)

	// create log group and stream
	err = jobDriver.CreateLogsInfra(ctx)
	require.Nil(t, err)

	// create dlq SQS for the mappers and coordinator
	dlqArn, err := jobDriver.CreateLambdaDLQ(ctx)
	require.Nil(t, err)

	// upload images to amazon ECR and create lambda function
	err = jobDriver.UploadLambdaFunctions(ctx, dlqArn)
	require.Nil(t, err)
}

func TestRunQ1(t *testing.T) {
	// currently ribble needs to be run at the root of the directory
	os.Chdir("../../../")

	jobID := uuid.MustParse("88cc574a-83b1-40fa-92fc-3b4d4fd24624")
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
	buildData, err := generators.ReadBuildData(jobID.String())
	require.Nil(t, err)
	jobDriver.BuildData = buildData

	// start coordinator
	err = jobDriver.StartCoordinator(ctx)
	require.Nil(t, err)

	// wait until job has completed
	assertOutputQ1(t, "./build/integration_tests/tests_expected_output/test1_out", jobID.String())
}
