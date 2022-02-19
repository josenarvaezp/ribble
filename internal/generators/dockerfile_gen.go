package generators

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/josenarvaezp/displ/internal/utils"
)

// DockerfileData defines the data needed for the template
type DockerfileData struct {
	JobID        string
	FunctionType string
	FunctionName string
	Workspace    string
}

// ExecuteDockerfileGenerator generates a dockerfile with the auto generated code
func ExecuteDockerfileGenerator(jobID string, workspace string) error {
	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s", GeneratedFilesDir, jobID)
	dockerfileDirName := fmt.Sprintf("%s/dockerfiles", generatedDirName)

	// create dir
	if _, err := os.Stat(dockerfileDirName); os.IsNotExist(err) {
		err := os.Mkdir(dockerfileDirName, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// each go file in the "lambda_gen" directory needs a dockerfile generated
	dirs, err := utils.ListFiles(generatedDirName)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		// skip dockerfiles directory
		if dir.Name() == "dockerfiles" {
			continue
		}

		currentDirName := fmt.Sprintf("%s/%s", generatedDirName, dir.Name())
		if dir.IsDir() {
			goFiles, err := utils.ListFiles(currentDirName)
			if err != nil {
				return err
			}

			// loop through files and generate a dockerfile for each
			for _, goFile := range goFiles {
				if !goFile.IsDir() {
					functionName := strings.SplitAfter(goFile.Name(), ".")[0]
					functionName = functionName[0 : len(functionName)-1]
					functionType := dir.Name()
					err = ExecuteSingleDockerfileGenerator(
						dockerfileDirName,
						jobID,
						workspace,
						functionType,
						functionName,
					)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// ExecuteSingleDockerfileGenerator generates a dockerfile
func ExecuteSingleDockerfileGenerator(
	dirName,
	jobID,
	workspace,
	functionType,
	functionName string,
) error {
	generatedFileName := fmt.Sprintf("%s/Dockerfile.%s", dirName, functionName)

	// create file
	f, err := os.OpenFile(generatedFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	// generate template
	t := template.Must(template.New("dockerfile").Parse(dockerfileTemplate))
	err = t.Execute(f, DockerfileData{
		JobID:        jobID,
		FunctionType: functionType,
		FunctionName: functionName,
		Workspace:    workspace,
	})
	return err
}
