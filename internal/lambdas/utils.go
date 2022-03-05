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
			"https://localstack:4566/000000000000/%s",
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
	switch aggregatorReflectType.Name() {
	case "MapAggregator":
		return MapAggregator
	case "Sum":
		return SumAggregator
	default:
		return InvalidAggregator
	}
}
