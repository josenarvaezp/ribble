package ribble

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/josenarvaezp/displ/internal/generators"
	"github.com/josenarvaezp/displ/pkg/aggregators"
	"gopkg.in/yaml.v2"
)

type Config struct {
	InputBuckets        []string `yaml:"input"`
	Region              string   `yaml:"region"`
	Local               bool     `yaml:"local"`
	LogLevel            int      `yaml:"logLevel"`
	AccountID           string   `yaml:"accountID"`
	Username            string   `yaml:"username"`
	LogicalSplit        bool     `yaml:"logicalSplit"`
	RandomizedPartition bool     `yaml:"randomizedPartition"`
}

func Job(
	mapper func(string) aggregators.MapAggregator,
	filter func(aggregators.MapAggregator) aggregators.MapAggregator,
	sort func(aggregators.MapAggregator) sort.Interface,
	config Config,
) error {
	// get job id and workspace from flags
	var workSpace string
	var jobID string

	flag.StringVar(&workSpace, "workspace", "", "The workspace for the job")
	flag.StringVar(&jobID, "job-id", "", "The ID for the job")
	flag.Parse()

	// validate mapper function
	err := generators.ValidateMapper(mapper)
	if err != nil {
		return err
	}

	// validate filter function
	if filter != nil {
		if err := generators.ValidateFilter(filter); err != nil {
			return err
		}
	}

	// validate sort function
	if sort != nil {
		if err := generators.ValidateSort(sort); err != nil {
			return err
		}
	}

	// get function name and package info
	mapperData := generators.GetFunctionData(mapper, jobID)

	// generate mapper file for lambda function
	err = generators.ExecuteMapperGenerator(jobID, config.RandomizedPartition, mapperData)
	if err != nil {
		return err
	}

	// generate coordinator
	coordinatorData := generators.GetCoordinatorData(jobID, mapperData)

	// generate coordinator file for lambda function
	err = generators.ExecuteCoordinatorGenerator(jobID, config.RandomizedPartition)
	if err != nil {
		return err
	}

	// get function name and package info
	reducerData := generators.GetReducerData(filter, sort, config.RandomizedPartition, jobID)

	// generate mapper file for lambda function
	err = generators.ExecuteReducerGenerator(jobID, config.RandomizedPartition, reducerData)
	if err != nil {
		return err
	}

	// generate coordinator dockerfile
	err = generators.ExecuteDockerfileGenerator(jobID, workSpace)
	if err != nil {
		return err
	}

	// write build data
	buildData := &generators.BuildData{
		JobPath:         workSpace,
		BuildDir:        fmt.Sprintf("%s/%s", generators.GeneratedFilesDir, jobID),
		MapperData:      mapperData,
		CoordinatorData: coordinatorData,
		ReducerData:     reducerData,
	}
	err = generators.WriteBuildData(buildData, jobID)
	if err != nil {
		return err
	}

	// write config fata
	err = writeConfigData(
		config,
		fmt.Sprintf("%s/%s", generators.GeneratedFilesDir, jobID),
		jobID,
	)
	if err != nil {
		return err
	}

	return nil
}

// writeConfigData writes the config data to a yaml file
func writeConfigData(config Config, generatedFilesDir string, jobID string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// create dir
	if _, err := os.Stat(generatedFilesDir); os.IsNotExist(err) {
		err := os.MkdirAll(generatedFilesDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// write yaml file
	fileName := fmt.Sprintf("%s/config.yaml", generatedFilesDir)
	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		return err
	}

	return err
}
