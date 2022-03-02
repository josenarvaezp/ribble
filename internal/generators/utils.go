package generators

import (
	"errors"

	"github.com/josenarvaezp/displ/internal/lambdas"
)

// AggregatorTypeToInternalFunction returns the string name of the
// function which computes the given aggregator
func AggregatorTypeToInternalFunction(aggregatorType lambdas.AggregatorType) (string, error) {
	switch aggregatorType {
	case lambdas.MapSumAggregator:
		return lambdas.ECRAggregatorMapSum, nil
	case lambdas.MapMaxAggregator:
		return lambdas.ECRAggregatorMapMax, nil
	case lambdas.MapMinAggregator:
		return lambdas.ECRAggregatorMapMin, nil
	default:
		return "", errors.New("Aggregator type is not valid")
	}
}
