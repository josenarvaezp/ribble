package generators

import (
	"errors"

	"github.com/josenarvaezp/displ/internal/lambdas"
)

// AggregatorTypeToCoordinatorData generates the data needed for the coordinator
// template according to the aggregator type
func AggregatorTypeToCoordinatorData(aggregatorType lambdas.AggregatorType) (*CoordinatorData, error) {
	coordinatorData := &CoordinatorData{}

	switch aggregatorType {
	case lambdas.MapAggregator:
		coordinatorData.LambdaAggregator = lambdas.ECRMapAggregator
	case lambdas.SumAggregator:
		coordinatorData.LambdaAggregator = lambdas.ECRMapAggregator
		// coordinatorData.LambdaFinalAggregator = lambdas.ECRAggregatorSumFinal
	default:
		return nil, errors.New("Aggregator type is not valid")
	}

	return coordinatorData, nil
}
