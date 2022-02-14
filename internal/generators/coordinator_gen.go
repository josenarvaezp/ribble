package generators

import (
	"fmt"
	"os"
	"text/template"

	"github.com/josenarvaezp/displ/internal/driver"
	"github.com/josenarvaezp/displ/pkg/aggregators"
)

// CoordinatorData defines the data needed for the template
type CoordinatorData struct {
	LambdaAggregator string
}

// ExecuteCoordinatorGenerator generates a go file with the auto generated code
func ExecuteCoordinatorGenerator(jobID string, aggregatorType aggregators.AggregatorType) error {
	// get string representation of the aggregator function
	internalAggregator, err := AggregatorTypeToInternalFunction(aggregatorType)
	if err != nil {
		return err
	}

	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s/coordinator", driver.GeneratedFilesDir, jobID)
	generatedFileName := fmt.Sprintf("%s/%s.go", generatedDirName, "coordinator")

	// create directory
	if _, err := os.Stat(generatedDirName); os.IsNotExist(err) {
		err := os.Mkdir(generatedDirName, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// create file
	fileWriter, err := os.OpenFile(generatedFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer fileWriter.Close()

	// generate template
	t := template.Must(template.New("coordinator").Parse(coordinatorTemplate))
	err = t.Execute(fileWriter, CoordinatorData{
		LambdaAggregator: internalAggregator,
	})
	return err
}
