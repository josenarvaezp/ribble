package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/josenarvaezp/displ/internal/driver"
	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/josenarvaezp/displ/internal/logs"
)

var (
	// Used for CLI flags
	jobPath string
	jobID   string
)

var (
	// vars used in init
	conf      *driver.Config
	jobDriver *driver.Driver
)

func init() {
	// get driver config values
	conf, err := driver.ReadLocalConfigFile("config.yaml")
	if err != nil {
		logrus.WithField(
			"File name", "TODO: get config file name",
		).WithError(err).Error("Error reading config file")
		return
	}

	// set logger
	logrus.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))

	// set driver
	jobID := uuid.New()
	jobDriver, err = driver.NewDriver(jobID, conf)
	if err != nil {
		logrus.WithError(err).Error("Error initializing driver")
		return
	}
}

func main() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(runCmd)

	buildCmd.PersistentFlags().StringVar(&jobPath, "job", "", "path to go file defining job")
	buildCmd.MarkPersistentFlagRequired("job")

	uploadCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to upload")
	uploadCmd.MarkPersistentFlagRequired("job-id")

	runCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to run")
	runCmd.MarkPersistentFlagRequired("job-id")

	rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ribble",
	Short: "Ribble is a distributed framework to process data in a serverless architecture",
	Long:  `Ribble is a distributed framework to process data in a serverless architecture`,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the resources needed for the processing job",
	Long:  `Build the resources needed for the processing job`,
	Run: func(cmd *cobra.Command, args []string) {
		// add job path info to driver
		jobDriver.BuildData = &generators.BuildData{
			JobPath:  jobPath,
			BuildDir: fmt.Sprintf("./build/lambda_gen/%s", jobDriver.JobID.String()),
		}

		// add loger info
		driverLogger := logrus.WithFields(logrus.Fields{
			"Job ID": jobDriver.JobID.String(),
		})

		// build directory for job's generated files
		if _, err := os.Stat(jobDriver.BuildData.BuildDir); os.IsNotExist(err) {
			err := os.MkdirAll(jobDriver.BuildData.BuildDir, os.ModePerm)
			if err != nil {
				driverLogger.WithError(err).Fatal("Error creating directory")
				return
			}
		}

		// build binary that creates lambda files
		err := jobDriver.BuildJobGenerationBinary()
		if err != nil {
			driverLogger.WithError(err).Fatal("Error building binary from job path")
			return
		}

		// run binary to create lambda files (go files and dockerfiles)
		fmt.Println("Generating resources...")
		err = jobDriver.GenerateResourcesFromBinary()
		if err != nil {
			driverLogger.WithError(err).Fatal("Error generating lambda files")
			return
		}

		// build mapper and coordinator docker images
		fmt.Println("Generating Images...")
		err = jobDriver.BuildDockerCustomImages()
		if err != nil {
			driverLogger.WithError(err).Fatal("Error building images")
			return
		}

		// build aggregator images
		err = jobDriver.BuildAggregatorImages()
		if err != nil {
			driverLogger.WithError(err).Fatal("Error building aggregator images")
			return
		}

		fmt.Println("Build successful with Job ID: ", jobDriver.JobID)
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload the resources needed for the processing job",
	Long:  `Upload the resources needed for the processing job`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// add job path info to driver
		jobID, err := uuid.Parse(jobID)
		if err != nil {
			logrus.WithError(err).Error("Error parsing ID, it must be an uuid")
			return
		}
		jobDriver.JobID = jobID

		// add loger info
		driverLogger := logrus.WithFields(logrus.Fields{
			"Job ID": jobDriver.JobID.String(),
		})

		// get build data
		buildData, err := generators.ReadBuildData(jobDriver.JobID.String())
		if err != nil {
			logrus.WithError(err).Error("Error reading build data")
			return
		}
		jobDriver.BuildData = buildData

		// upload images to amazon ECR
		err = jobDriver.UploadJobImages(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error uploading images")
			return
		}

		// Setting up resources
		err = jobDriver.CreateJobBucket(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating the job bucket")
			return
		}

		// create streams for job
		err = jobDriver.CreateQueues(ctx, 5) // TODO: get num of queues
		if err != nil {
			driverLogger.WithError(err).Error("Error creating the job streams")
			return
		}

		fmt.Println("Upload successful with Job ID: ", jobDriver.JobID)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the job",
	Long:  `Run the job`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// add job path info to driver
		jobID, err := uuid.Parse(jobID)
		if err != nil {
			logrus.WithError(err).Error("Error parsing ID, it must be an uuid")
			return
		}
		jobDriver.JobID = jobID

		driverLogger := logrus.WithFields(logrus.Fields{
			"Job ID": jobID.String(),
		})

		// get build data
		buildData, err := generators.ReadBuildData(jobDriver.JobID.String())
		if err != nil {
			logrus.WithError(err).Error("Error reading build data")
			return
		}
		jobDriver.BuildData = buildData

		// generate mappings from S3 input bucket
		mappings, err := jobDriver.GenerateMappingsCompleteObjects(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error generating mappings from S3")
			return
		}

		// start coordinator
		err = jobDriver.StartCoordinator(ctx, len(mappings), 5 /*TODO: num queues*/)
		if err != nil {
			driverLogger.WithError(err).Error("Error starting the coordinator")
			return
		}

		// start mappers
		err = jobDriver.StartMappers(ctx, mappings)
		if err != nil {
			driverLogger.WithError(err).Error("Error starting the mappers")
			return
		}

		fmt.Println("Running job: ", jobDriver.JobID)
	},
}
