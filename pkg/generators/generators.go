package generators

import (
	"flag"

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
	functionData := generators.GetFunctionData(mapper)

	// generate mapper file for lambda function
	err = generators.ExecuteMapperGenerator(jobID, aggregatorType, functionData)
	if err != nil {
		return err
	}

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

	return nil
}
