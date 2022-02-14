package generators

import (
	"fmt"
	"os"
	"text/template"

	"github.com/josenarvaezp/displ/internal/driver"
)

// ExecuteMapSumGenerator generates a go file with the auto generated code
// for the corresponding mapper function
func ExecuteMapSumGenerator(jobID string, data FunctionData) error {
	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s/map_sum", driver.GeneratedFilesDir, jobID)
	generatedFileName := fmt.Sprintf("%s/%s.go", generatedDirName, data.Function)

	// create dir
	if _, err := os.Stat(generatedDirName); os.IsNotExist(err) {
		err := os.MkdirAll(generatedDirName, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// create file
	f, err := os.OpenFile(generatedFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	// generate file with template
	t := template.Must(template.New("mapper").Parse(mapTemplate))
	err = t.Execute(f, data)

	return err
}
