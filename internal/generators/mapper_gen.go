package generators

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

var (
	// vars used to compare user data type
	stringType        = reflect.TypeOf("")
	mapAggregatorType = reflect.TypeOf(make(aggregators.MapAggregator))
)

// FunctionData defines the data needed for the template
type FunctionData struct {
	PackagePath   string `yaml:"PackagePath,omitempty"`
	PackageName   string `yaml:"PackageName,omitempty"`
	GeneratedFile string `yaml:"GeneratedFile,omitempty"`
	Function      string `yaml:"Function,omitempty"`
	ImageName     string `yaml:"ImageName,omitempty"`
	ImageTag      string `yaml:"ImageTag,omitempty"`
	Dockefile     string `yaml:"Dockerfile,omitempty"`
	Aggregator    string `yaml:"Aggregator,omitempty"`
}

// GetFunctionData gets as input an interface that should be a function
// and gets the function's package information and the function name
func GetFunctionData(i interface{}, jobID string) *FunctionData {
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	lenFullName := len(fullName)

	// used to get the last package.file.go
	indexOfLastSlash := strings.LastIndex(fullName, "/")
	// used to get file.go
	indexOfSecondLastDot := strings.LastIndex(fullName[0:lenFullName-3], ".")

	packageName := fullName[indexOfLastSlash+1 : indexOfSecondLastDot]
	packageFullName := fullName[0:indexOfLastSlash+1] + packageName
	functionName := fullName[indexOfSecondLastDot+1 : lenFullName]

	return &FunctionData{
		PackagePath: packageFullName,
		PackageName: packageName,
		GeneratedFile: fmt.Sprintf("%s/%s/%s/%s.go",
			GeneratedFilesDir,
			jobID,
			"map",
			functionName,
		),
		Function:  functionName,
		ImageName: fmt.Sprintf("%s_%s", strings.ToLower(functionName), jobID),
		ImageTag:  "latest",
		Dockefile: fmt.Sprintf("%s/%s/dockerfiles/Dockerfile.%s",
			GeneratedFilesDir,
			jobID,
			functionName,
		),
	}
}

// ValidateMapper gets a mapper function as input and check that its
// return type is a valid aggregator type
func ValidateMapper(mapper interface{}) error {
	mapperType := reflect.TypeOf(mapper)

	// validate that the input of the function gets one argument
	// which should be the filename
	if mapperType.NumIn() != 1 {
		return errors.New("Invalid error signature. The mapper function can only take the filename as input")
	}

	// validate the input of function is a string
	if !mapperType.In(0).ConvertibleTo(stringType) {
		return errors.New("Invalid error signature. The input to the function should be a string")
	}

	// validate that the function returns a single value
	if mapperType.NumOut() != 1 {
		return errors.New("The mapper function can only have one output")
	}

	// return the aggregator specified or return error if the output
	// is not a valid aggregator
	aggregatorType := mapperType.Out(0)

	switch aggregatorType.Name() {
	case "MapAggregator":
		return nil
	default:
		return errors.New("Invalid aggregator used")
	}
}

// ExecuteMapGenerator generates a go file with the auto generated code
// for the corresponding mapper function
func ExecuteMapGenerator(jobID string, data *FunctionData, functionTemplate string) error {
	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s/map", GeneratedFilesDir, jobID)
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
	t := template.Must(template.New("mapper").Parse(functionTemplate))
	err = t.Execute(f, data)
	return err
}

// ExecuteMapperGenerator generates code for the mapper according
// to the aggregator used
func ExecuteMapperGenerator(
	jobID string,
	randomizedPartition bool,
	functionData *FunctionData,
) error {

	var template string

	if randomizedPartition {
		template = mapRandomTemplate
	} else {
		template = mapTemplate
	}

	err := ExecuteMapGenerator(jobID, functionData, template)
	if err != nil {
		return err
	}

	return nil
}
