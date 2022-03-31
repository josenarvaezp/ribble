package generators

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	// name of directories and flags for binary
	PathToRoot                    = "."
	BinaryNameToBuildJob          = "gen_job"
	WorkspaceFlag                 = "--workspace"
	JobIdFlag                     = "--job-id"
	JobLocalFlag                  = "--local"
	ScriptToGenerateGoFiles       = "./build/generate_lambda_files.sh"
	ScriptToBuildImages           = "./build/build_dockerfiles.sh"
	ScriptToBuildAggregatorImages = "./build/build_aggregators.sh"
	GeneratedFilesDir             = "./build/lambda_gen"
	AggregatorDockerfiles         = "./build/aggregators"
	BuildDataFile                 = "build.yaml"
)

type BuildData struct {
	JobPath         string                 `yaml:"JobPath,omitempty"`
	BuildDir        string                 `yaml:"BuildDir,omitempty"`
	MapperData      *FunctionData          `yaml:"MapperData,omitempty"`
	CoordinatorData *CoordinatorData       `yaml:"CoordinatorData,omitempty"`
	ReducerData     []*ReducerFunctionData `yaml:"ReducerData,omitempty"`
}

// WriteBuildData writes the build data to a yaml file
func WriteBuildData(buildData *BuildData, jobID string) error {
	data, err := yaml.Marshal(buildData)
	if err != nil {
		return err
	}

	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s", GeneratedFilesDir, jobID)

	// create dir
	if _, err := os.Stat(generatedDirName); os.IsNotExist(err) {
		err := os.MkdirAll(generatedDirName, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// write yaml file
	fileName := fmt.Sprintf("%s/build.yaml", generatedDirName)
	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		return err
	}

	return err
}

// ReadBuildData reads the build data from a yaml file
func ReadBuildData(jobID string) (*BuildData, error) {
	buildFile := fmt.Sprintf("%s/%s/%s", GeneratedFilesDir, jobID, BuildDataFile)
	yamlFile, err := ioutil.ReadFile(buildFile)
	if err != nil {
		return nil, err
	}

	var buildData BuildData
	err = yaml.Unmarshal(yamlFile, &buildData)
	if err != nil {
		return nil, err
	}

	return &buildData, nil
}
