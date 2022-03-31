package generators

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"github.com/josenarvaezp/displ/internal/lambdas"
)

// ReducerFunctionData defines the data needed for the template of the reducers
type ReducerFunctionData struct {
	ReducerName    string `yaml:"ReducerName,omitempty"`
	PackagePath    string `yaml:"PackagePath,omitempty"`
	PackageName    string `yaml:"PackageName,omitempty"`
	GeneratedFile  string `yaml:"GeneratedFile,omitempty"`
	FilterFunction string `yaml:"FilterFunction,omitempty"`
	SortFunction   string `yaml:"SortFunction,omitempty"`
	ImageName      string `yaml:"ImageName,omitempty"`
	ImageTag       string `yaml:"ImageTag,omitempty"`
	Dockefile      string `yaml:"Dockerfile,omitempty"`
	Local          bool   `yaml:"Local,omitempty"`
}

// GetReducerData gets as input an interface that should be a function
// and gets the function's package information and the function name
func GetReducerData(filter interface{}, sort interface{}, randomizedPartition bool, jobID string, local bool) []*ReducerFunctionData {
	functionData := []*ReducerFunctionData{}
	if randomizedPartition {
		// add final and random reducers data
		functionData = append(functionData, &ReducerFunctionData{
			ReducerName: lambdas.ECRFinalMapAggregator,
			Local:       local,
		})

		functionData = append(functionData, &ReducerFunctionData{
			ReducerName: lambdas.ECRRandomMapAggregator,
			Local:       local,
		})

		for i := 0; i < 2; i++ {
			functionData[i].GeneratedFile = fmt.Sprintf("%s/%s/%s/%s.go",
				GeneratedFilesDir,
				jobID,
				functionData[i].ReducerName,
				functionData[i].ReducerName,
			)

			functionData[i].ImageName = fmt.Sprintf("%s_%s", strings.ToLower(functionData[i].ReducerName), jobID)
			functionData[i].ImageTag = "latest"
			functionData[i].Dockefile = fmt.Sprintf("%s/%s/dockerfiles/Dockerfile.%s",
				GeneratedFilesDir,
				jobID,
				functionData[i].ReducerName,
			)
		}
	} else {
		functionData = append(functionData, &ReducerFunctionData{
			ReducerName: lambdas.ECRMapAggregator,
			Local:       local,
		})
		functionData[0].GeneratedFile = fmt.Sprintf("%s/%s/%s/%s.go",
			GeneratedFilesDir,
			jobID,
			lambdas.ECRMapAggregator,
			lambdas.ECRMapAggregator,
		)

		functionData[0].ImageName = fmt.Sprintf("%s_%s", lambdas.ECRMapAggregator, jobID)
		functionData[0].ImageTag = "latest"
		functionData[0].Dockefile = fmt.Sprintf("%s/%s/dockerfiles/Dockerfile.%s",
			GeneratedFilesDir,
			jobID,
			lambdas.ECRMapAggregator,
		)
	}

	if filter != nil {
		// get info for filter
		filterFullName := runtime.FuncForPC(reflect.ValueOf(filter).Pointer()).Name()
		filterLenFullName := len(filterFullName)

		// used to get the last package.file.go
		filterIndexOfLastSlash := strings.LastIndex(filterFullName, "/")
		// used to get file.go
		filterIndexOfSecondLastDot := strings.LastIndex(filterFullName[0:filterLenFullName-3], ".")

		functionData[0].FilterFunction = filterFullName[filterIndexOfSecondLastDot+1 : filterLenFullName]

		packageName := filterFullName[filterIndexOfLastSlash+1 : filterIndexOfSecondLastDot]
		packagePath := filterFullName[0:filterIndexOfLastSlash+1] + packageName

		functionData[0].PackagePath = packagePath
		functionData[0].PackageName = packageName
	}

	if sort != nil {
		// get info for sorting
		sortFullName := runtime.FuncForPC(reflect.ValueOf(sort).Pointer()).Name()
		sortLenFullName := len(sortFullName)

		// used to get the last package.file.go
		sortIndexOfLastSlash := strings.LastIndex(sortFullName, "/")
		// used to get file.go
		sortIndexOfSecondLastDot := strings.LastIndex(sortFullName[0:sortLenFullName-3], ".")

		functionData[0].SortFunction = sortFullName[sortIndexOfSecondLastDot+1 : sortLenFullName]

		packageName := sortFullName[sortIndexOfLastSlash+1 : sortIndexOfSecondLastDot]
		packagePath := sortFullName[0:sortIndexOfLastSlash+1] + packageName

		functionData[0].PackagePath = packagePath
		functionData[0].PackageName = packageName
	}

	return functionData
}

// ValidateFilter gets a filter function as input and check that its
// return type is valid
func ValidateFilter(filterFunc interface{}) error {
	filterType := reflect.TypeOf(filterFunc)

	// validate that the input of the function gets one argument
	// which should be a map aggregator
	if filterType.NumIn() != 1 {
		return errors.New("Invalid error signature. The filter function can only take the map aggregator as input")
	}

	// validate the input of function is a map aggregator
	if !filterType.In(0).ConvertibleTo(mapAggregatorType) {
		return errors.New("Invalid error signature. The input to the function should be a map aggregator")
	}

	// validate that the function returns a single value
	if filterType.NumOut() != 1 {
		return errors.New("The filter function can only have one output")
	}

	// return the aggregator specified or return error if the output
	// is not a valid aggregator
	aggregatorType := filterType.Out(0)

	switch aggregatorType.Name() {
	case "MapAggregator":
		return nil
	default:
		return errors.New("Invalid aggregator output")
	}
}

// ValidateSort gets a sort function as input and check that its
// return type is valid
func ValidateSort(sortFunc interface{}) error {
	sortType := reflect.TypeOf(sortFunc)

	// validate that the input of the function gets one argument
	// which should be a map aggregator
	if sortType.NumIn() != 1 {
		return errors.New("Invalid error signature. The sort function can only take the map aggregator as input")
	}

	// validate the input of function is a map aggregator
	if !sortType.In(0).ConvertibleTo(mapAggregatorType) {
		return errors.New("Invalid error signature. The input to the function should be a map aggregator")
	}

	// validate that the function returns a single value
	if sortType.NumOut() != 1 {
		return errors.New("The filter function can only have one output")
	}

	// check the output implements the sort.Interface
	output := sortType.Out(0)
	if output.Name() != "Interface" || output.PkgPath() != "sort" {
		return errors.New("Invalid sort output")
	}

	return nil
}

// ExecuteReduceGenerator generates a go file with the auto generated code
// for the corresponding reducer function
func ExecuteReduceGenerator(jobID string, data *ReducerFunctionData, functionTemplate string) error {
	// dir and file where generated code is writen to
	generatedDirName := fmt.Sprintf("%s/%s/%s", GeneratedFilesDir, jobID, data.ReducerName)
	generatedFileName := fmt.Sprintf("%s/%s.go", generatedDirName, data.ReducerName)

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
	t := template.Must(template.New("reducer").Parse(functionTemplate))
	if err = t.Execute(f, data); err != nil {
		return err
	}

	return nil
}

// ExecuteReducerGenerator generates code for the reducers
func ExecuteReducerGenerator(
	jobID string,
	randomizedPartition bool,
	functionData []*ReducerFunctionData,
) error {
	if randomizedPartition {
		// create aggregators for randomized partition
		if err := ExecuteReduceGenerator(jobID, functionData[0], reduceMapFinalAggregatorTemplate); err != nil {
			return err
		}

		if err := ExecuteReduceGenerator(jobID, functionData[1], reduceRandomMapAggregatorTemplate); err != nil {
			return err
		}
	} else {
		// create aggregators
		if err := ExecuteReduceGenerator(jobID, functionData[0], reduceMapAggregatorTemplate); err != nil {
			return err
		}
	}

	return nil
}
