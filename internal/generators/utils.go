package generators

import (
	"errors"

	"github.com/josenarvaezp/displ/internal/lambdas"
)

const (
	internalMapSumAggregator = "lambdas.ECRAggregatorMapSum"
)

// AggregatorTypeToInternalFunction returns the string name of the
// function which computes the given aggregator
func AggregatorTypeToInternalFunction(aggregatorType lambdas.AggregatorType) (string, error) {
	switch aggregatorType {
	case lambdas.MapSumAggregator:
		return internalMapSumAggregator, nil
	case lambdas.SumAggregator:
		return "TODO", errors.New("unimplemented")
	default:
		return "", errors.New("Aggregator type is not valid")
	}
}
