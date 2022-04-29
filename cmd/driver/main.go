package main

import (
	"context"
	"fmt"
	"math"
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
	jobPath   string
	jobID     string
	accountID string
	username  string
	region    string
	verbose   *int
	local     bool
	logsSleep int32
	reducers  int
)

func main() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setCredsCmd)
	rootCmd.AddCommand(logsCmd)

	setCredsCmd.PersistentFlags().StringVar(&accountID, "account-id", "", "AWS account id")
	setCredsCmd.PersistentFlags().StringVar(&username, "username", "", "AWS username")
	setCredsCmd.PersistentFlags().BoolVar(&local, "local", false, "local environment")
	setCredsCmd.MarkFlagRequired("account-id")
	setCredsCmd.MarkFlagRequired("username")
	setCredsCmd.Flags().CountP("verbose", "v", "counted verbosity")

	buildCmd.PersistentFlags().StringVar(&jobPath, "job", "", "path to go file defining job")
	buildCmd.MarkPersistentFlagRequired("job")
	buildCmd.Flags().CountP("verbose", "v", "counted verbosity")

	uploadCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to upload")
	uploadCmd.PersistentFlags().IntVar(&reducers, "reducers", 0, "number of reducers to use")
	uploadCmd.MarkPersistentFlagRequired("job-id")
	uploadCmd.Flags().CountP("verbose", "v", "counted verbosity")

	runCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to run")
	runCmd.MarkPersistentFlagRequired("job-id")
	runCmd.Flags().CountP("verbose", "v", "counted verbosity")

	logsCmd.PersistentFlags().StringVar(&jobID, "job-id", "", "id of job to run")
	logsCmd.PersistentFlags().Int32Var(&logsSleep, "sleep", 60, "time in seconds for fetching logs")
	logsCmd.MarkPersistentFlagRequired("job-id")

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

		// get verbosity for logs
		verbosity, _ := cmd.Flags().GetCount("verbose")
		logrus.SetLevel(logs.ConfigLogLevelToLevel(verbosity))

		// set driver
		jobID := uuid.New()
		jobDriver := driver.NewBuildDriver(jobID)

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
		fmt.Println("Building docker images...")
		err = jobDriver.BuildDockerCustomImages()
		if err != nil {
			driverLogger.WithError(err).Fatal("Error building images")
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
		fmt.Println("Creating resources...")

		ctx := context.Background()

		// get verbosity for logs
		verbosity, _ := cmd.Flags().GetCount("verbose")
		logrus.SetLevel(logs.ConfigLogLevelToLevel(verbosity))

		// get driver config values
		configFile := fmt.Sprintf("%s/%s/config.yaml", generators.GeneratedFilesDir, jobID)
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

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

		// Setting up resources
		err = jobDriver.CreateJobBucket(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating the job bucket")
			return
		}

		// generate mappings from S3 input bucket
		mappings, err := jobDriver.GenerateMappings(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error generating mappings from S3")
			return
		}

		// get number of reducers
		numMappings := len(mappings)
		if reducers == 0 {
			// no reducers specified
			reducers = int(math.Ceil(float64(numMappings) / 2))
		}

		// update build data
		buildData.NumMappers = numMappings
		buildData.NumReducers = reducers
		err = generators.WriteBuildData(buildData, jobID.String())
		if err != nil {
			driverLogger.WithError(err).Error("Error updating build data")
			return
		}

		// write mappings to s3
		err = jobDriver.WriteMappings(ctx, mappings)
		if err != nil {
			driverLogger.WithError(err).Error("Error writing mappings to S3")
			return
		}

		// create streams for job
		err = jobDriver.CreateQueues(ctx, reducers)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating the job streams")
			return
		}

		// create log group and stream
		err = jobDriver.CreateLogsInfra(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating log group and stream")
			return
		}

		// create dlq SQS for the mappers and coordinator
		dlqArn, err := jobDriver.CreateLambdaDLQ(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating the dead-letter queue for the job mappers")
			return
		}

		// upload images to amazon ECR and create lambda function
		err = jobDriver.UploadLambdaFunctions(ctx, dlqArn)
		if err != nil {
			driverLogger.WithError(err).Error("Error creating functions")
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

		// get verbosity for logs
		verbosity, _ := cmd.Flags().GetCount("verbose")
		logrus.SetLevel(logs.ConfigLogLevelToLevel(verbosity))

		// get driver config values
		configFile := fmt.Sprintf("%s/%s/config.yaml", generators.GeneratedFilesDir, jobID)
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

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

		// start coordinator
		err = jobDriver.StartCoordinator(ctx)
		if err != nil {
			driverLogger.WithError(err).Error("Error starting the coordinator")
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

		// get verbosity for logs
		verbosity, _ := cmd.Flags().GetCount("verbose")
		logrus.SetLevel(logs.ConfigLogLevelToLevel(verbosity))

		// set driver
		jobDriver, err := driver.NewSetupDriver(&config.Config{
			AccountID: accountID,
			Username:  username,
		})
		if err != nil {
			logrus.WithError(err).Error("Error initializing driver")
			return
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

		fmt.Println("Ribble credentials created succesfully")
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get job logs",
	Long:  `Get job logs`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// get driver config values
		configFile := fmt.Sprintf("%s/%s/config.yaml", generators.GeneratedFilesDir, jobID)
		conf, err := config.ReadLocalConfigFile(configFile)
		if err != nil {
			logrus.WithField(
				"File name", configFile,
			).WithError(err).Error("Error reading config file")
			return
		}

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

		err = jobDriver.ReadRibbleLogs(ctx, logsSleep)
		if err != nil {
			logrus.WithError(err).Error("Error reading logs")
			return
		}
	},
}
