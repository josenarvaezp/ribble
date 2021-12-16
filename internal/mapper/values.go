package mapper

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
