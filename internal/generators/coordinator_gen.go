package generators

import (
	"fmt"
	"os"
	"text/template"

	"github.com/josenarvaezp/displ/pkg/lambdas"
)

// CoordinatorData defines the data needed for the template
type CoordinatorData struct {
	LambdaAggregator      string `yaml:"LambdaAggregator,omitempty"`
	LambdaFinalAggregator string `yaml:"LambdaFinalAggregator,omitempty"`
	GeneratedFile         string `yaml:"GeneratedFile,omitempty"`
	Function              string `yaml:"Function,omitempty"`
	ImageName             string `yaml:"ImageName,omitempty"`
	ImageTag              string `yaml:"ImageTag,omitempty"`
	Dockefile             string `yaml:"Dockerfile,omitempty"`
	Local                 bool   `yaml:"Local,omitempty"`
}

func GetCoordinatorData(jobID string, mapperData *FunctionData, randomizedPartition, local bool) *CoordinatorData {
	coordinatorData := &CoordinatorData{
		Function: "coordinator",
		GeneratedFile: fmt.Sprintf("%s/%s/%s/%s.go",
			GeneratedFilesDir,
			jobID,
			"coordinator",
			"coordinator",
		),
		ImageName: fmt.Sprintf("%s_%s", "coordinator", jobID),
		ImageTag:  "latest",
		Dockefile: fmt.Sprintf("%s/%s/dockerfiles/Dockerfile.coordinator",
			GeneratedFilesDir,
			jobID,
		),
		Local: local,
	}

	// get correct coordinator data for template
	if randomizedPartition {
		coordinatorData.LambdaAggregator = lambdas.ECRRandomMapAggregator
		coordinatorData.LambdaFinalAggregator = lambdas.ECRFinalMapAggregator
	} else {
		coordinatorData.LambdaAggregator = lambdas.ECRMapAggregator
	}

	return coordinatorData
}

// ExecuteCoordinatorGenerator generates a go file with the auto generated code
func ExecuteCoordinatorGenerator(jobID string, randomizedPartition bool, coordinatorData *CoordinatorData) error {
	// get correct template
	var templateValue string
	if randomizedPartition {
		templateValue = randomCoordinatorTemplate
	} else {
		templateValue = mapCoordinatorTemplate
	}

	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s/coordinator", GeneratedFilesDir, jobID)
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
	t := template.Must(template.New("coordinator").Parse(templateValue))
	err = t.Execute(fileWriter, coordinatorData)
	return err
}
