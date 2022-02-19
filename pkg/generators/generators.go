package generators

import (
	"flag"
	"fmt"

	"github.com/josenarvaezp/displ/internal/generators"
)

func Job(mapper interface{}) error {
	// get job id and workspace from flags
	var workSpace string
	var jobID string

	flag.StringVar(&workSpace, "workspace", "", "The workspace for the job")
	flag.StringVar(&jobID, "job-id", "", "The ID for the job")
	flag.Parse()

	// validate mapper function
	aggregatorType, err := generators.ValidateMapper(mapper)
	if err != nil {
		return err
	}

	// get function name and package info
	mapperData := generators.GetFunctionData(mapper, jobID)

	// generate mapper file for lambda function
	err = generators.ExecuteMapperGenerator(jobID, aggregatorType, mapperData)
	if err != nil {
		return err
	}

	// generate coordinator
	coordinatorData := generators.GetCoordinatorData(jobID, mapperData)

	// generate coordinator file for lambda function
	err = generators.ExecuteCoordinatorGenerator(jobID, aggregatorType)
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
	}
	err = generators.WriteBuildData(buildData, jobID)
	if err != nil {
		return err
	}

	return nil
}
