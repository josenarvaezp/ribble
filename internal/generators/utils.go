package generators

import (
	"errors"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

const (
	internalMapSumAggregator = "aggregators.AggregatorMapSum"
)

// AggregatorTypeToInternalFunction returns the string name of the
// function which computes the given aggregator
func AggregatorTypeToInternalFunction(aggregatorType aggregators.AggregatorType) (string, error) {
	switch aggregatorType {
	case aggregators.MapSumAggregator:
		return internalMapSumAggregator, nil
	case aggregators.SumAggregator:
		return "TODO", errors.New("unimplemented")
	default:
		return "", errors.New("Aggregator type is not valid")
	}
}
