package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/driver"
	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/josenarvaezp/displ/internal/logs"
)

var (
	// Used for CLI flags
	jobPath    string
	jobID      string
	accountID  string
	username   string
	configFile string
)

func main() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setCredsCmd)

	setCredsCmd.PersistentFlags().StringVar(&accountID, "account-id", "", "AWS account id")
	setCredsCmd.PersistentFlags().StringVar(&username, "username", "", "AWS username")
	setCredsCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	setCredsCmd.MarkFlagRequired("account-id")
	setCredsCmd.MarkFlagRequired("username")
	setCredsCmd.MarkPersistentFlagRequired("config")

	buildCmd.PersistentFlags().StringVar(&jobPath, "job", "", "path to go file defining job")
	buildCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	buildCmd.MarkPersistentFlagRequired("job")
	buildCmd.MarkPersistentFlagRequired("config")

	uploadCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to upload")
	uploadCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	uploadCmd.MarkPersistentFlagRequired("job-id")
	uploadCmd.MarkPersistentFlagRequired("config")

	runCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to run")
	runCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	runCmd.MarkPersistentFlagRequired("job-id")
	runCmd.MarkPersistentFlagRequired("config")

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

		// get driver config values
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

		// set logger
		logrus.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))

		// set driver
		jobID := uuid.New()
		jobDriver, err := driver.NewDriver(jobID, conf)
		if err != nil {
			logrus.WithError(err).Error("Error initializing driver")
			return
		}

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
		err = jobDriver.BuildJobGenerationBinary()
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

		// get driver config values
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

		// set logger
		logrus.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))

		// add job path info to driver
		jobID, err := uuid.Parse(jobID)
		if err != nil {
			logrus.WithError(err).Error("Error parsing ID, it must be an uuid")
			return
		}

		// set driver
		jobDriver, err := driver.NewDriver(jobID, conf)
		if err != nil {
			logrus.WithError(err).Error("Error initializing driver")
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

		// get driver config values
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

		// set logger
		logrus.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))

		// add job path info to driver
		jobID, err := uuid.Parse(jobID)
		if err != nil {
			logrus.WithError(err).Error("Error parsing ID, it must be an uuid")
			return
		}

		// set driver
		jobDriver, err := driver.NewDriver(jobID, conf)
		if err != nil {
			logrus.WithError(err).Error("Error initializing driver")
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

var setCredsCmd = &cobra.Command{
	Use:   "set-credentials",
	Short: "Set credentials for the job",
	Long:  `Set credentials for the job`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// get driver config values
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

		// set logger
		logrus.SetLevel(logs.ConfigLogLevelToLevel(conf.LogLevel))

		// set driver
		jobDriver, err := driver.NewSetupDriver(conf)
		if err != nil {
			logrus.WithError(err).Error("Error initializing driver")
			return
		}

		// if flags are used for username and account id override
		// the information fetched from the config yaml
		if username != "" {
			jobDriver.Config.Username = username
		}

		if accountID != "" {
			jobDriver.Config.AccountID = accountID
		}

		// create role
		_, err = jobDriver.CreateRole(ctx)
		if err != nil {
			logrus.WithError(err).Error("Error creating ribble role")
			return
		}

		// create policy
		policyARN, err := jobDriver.CreateRolePolicy(ctx)
		if err != nil {
			logrus.WithError(err).Error("Error creating ribble policy")
			return
		}

		// attach policy to role
		err = jobDriver.AttachRolePolicy(ctx, policyARN)
		if err != nil {
			logrus.WithError(err).Error("Error attaching ribber policy to role")
			return
		}

		// create policy to allow user to use role
		userPolicyArn, err := jobDriver.CreateUserPolicy(ctx)
		if err != nil {
			logrus.WithError(err).Error("Error creating user policy to assume ribble role")
			return
		}

		// attach policy to user
		err = jobDriver.AttachUserPolicy(ctx, userPolicyArn)
		if err != nil {
			logrus.WithError(err).Error("Error attaching policy to user")
			return
		}

		fmt.Println("IAM resources created succesfully")
	},
}
