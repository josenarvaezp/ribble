package lambdas

import (
	"fmt"
	"reflect"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

const (
	// items per batch
	MaxItemsPerBatch = 10

	// attributes for sending and receiving messages
	MapIDAttribute     = "map-id"
	BatchIDAttribute   = "batch-id"
	MessageIDAttribute = "message-id"
)

var (
	// MessageAttributes values
	numberDataType = "Number"
	stringDataType = "String"
)

// getQueueURL returns the queue URL based on its name
func GetQueueURL(queueName string, region string, accountID string, local bool) string {
	var queueURL string

	if local {
		queueURL = fmt.Sprintf(
			"https://localhost:4566/000000000000/%s",
			queueName,
		)
	} else {
		queueURL = fmt.Sprintf(
			"https://sqs.%s.amazonaws.com/%s/%s",
			region,
			accountID,
			queueName,
		)
	}

	return queueURL
}

func GetAggregatorType(value aggregators.Aggregator) AggregatorType {
	aggregatorReflectType := reflect.TypeOf(value)

	mapType := reflect.TypeOf(make(aggregators.MapAggregator))
	if aggregatorReflectType.ConvertibleTo(mapType) {
		return MapAggregator
	}

	sumType := reflect.TypeOf(aggregators.InitSum(0))
	if aggregatorReflectType.ConvertibleTo(sumType) {
		return SumAggregator
	}

	maxType := reflect.TypeOf(aggregators.InitMax(0))
	if aggregatorReflectType.ConvertibleTo(maxType) {
		return MaxAggregator
	}

	minType := reflect.TypeOf(aggregators.InitMin(0))
	if aggregatorReflectType.ConvertibleTo(minType) {
		return MinAggregator
	}

	avgType := reflect.TypeOf(aggregators.InitAvg(0, 0))
	if aggregatorReflectType.ConvertibleTo(avgType) {
		return AvgAggregator
	}

	return InvalidAggregator
}
