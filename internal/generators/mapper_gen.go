package generators

import (
	"errors"
	"reflect"
	"runtime"
	"strings"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

var (
	// vars used to compare user data type
	mapSumType = reflect.TypeOf(aggregators.MapSum{})
	sumType    = reflect.TypeOf(aggregators.Sum(0))
	stringType = reflect.TypeOf("")
)

// FunctionData defines the data needed for the template
type FunctionData struct {
	PackagePath string
	PackageName string
	Function    string
}

// GetFunctionData gets as input an interface that should be a function
// and gets the function's package information and the function name
func GetFunctionData(i interface{}) FunctionData {
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	lenFullName := len(fullName)

	// used to get the last package.file.go
	indexOfLastSlash := strings.LastIndex(fullName, "/")
	// used to get file.go
	indexOfSecondLastDot := strings.LastIndex(fullName[0:lenFullName-3], ".")

	packageName := fullName[indexOfLastSlash+1 : indexOfSecondLastDot]
	packageFullName := fullName[0:indexOfLastSlash+1] + packageName
	functionName := fullName[indexOfSecondLastDot+1 : lenFullName]

	return FunctionData{
		PackagePath: packageFullName,
		PackageName: packageName,
		Function:    functionName,
	}
}

// ValidateMapper gets a mapper function as input and check that its
// return type is a valid aggregator type
func ValidateMapper(mapper interface{}) (aggregators.AggregatorType, error) {
	mapperType := reflect.TypeOf(mapper)

	// validate that the input of the function gets one argument
	// which should be the filename
	if mapperType.NumIn() != 1 {
		return aggregators.InvalidAggregator,
			errors.New("Invalid error signature. The mapper function can only take the filename as input")
	}

	// validate the input of function is a string
	if !mapperType.In(0).ConvertibleTo(stringType) {
		return aggregators.InvalidAggregator,
			errors.New("Invalid error signature. The input to the function should be a string")
	}

	// validate that the function returns a single value
	if mapperType.NumOut() != 1 {
		return aggregators.InvalidAggregator, errors.New("The mapper function can only have one output")
	}

	// return the aggregator specified or return error if the output
	// is not a valid aggregator
	aggregatorType := mapperType.Out(0)

	if aggregatorType.ConvertibleTo(mapSumType) {
		return aggregators.MapSumAggregator, nil
	}

	if aggregatorType.ConvertibleTo(sumType) {
		return aggregators.SumAggregator, nil
	}

	return aggregators.InvalidAggregator, errors.New("Aggregator is invalid")
}

// ExecuteMapperGenerator generates code for the mapper according
// to the aggregator used
func ExecuteMapperGenerator(
	jobID string,
	aggregatorType aggregators.AggregatorType,
	functionData FunctionData,
) error {
	switch aggregatorType {
	case aggregators.MapSumAggregator:
		err := ExecuteMapSumGenerator(jobID, functionData)
		if err != nil {
			return err
		}
	case aggregators.SumAggregator:
		return errors.New("Unimplemented")
	default:
		return errors.New("Invalid aggregator")
	}

	return nil
}
