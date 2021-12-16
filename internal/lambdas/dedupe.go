package lambdas

// DedupeMap is a map holding the values of the already processed
// messages. The key to the first map represents the mappers' IDs.
// The second key represents the batchID for the current mapper.
type DedupeMap map[string]map[int]*DedupeProcessedMessages

// InitDedupeMap initializes a dedupe map
func InitDedupeMap() DedupeMap {
	return make(map[string]map[int]*DedupeProcessedMessages)
}

// InitDedupeBatch initializes the map for the current map and batch
func (d DedupeMap) InitDedupeBatch(mapID string, batchID int, messageID int) {
	d[mapID] = make(map[int]*DedupeProcessedMessages)
	d[mapID][batchID] = &DedupeProcessedMessages{
		ProcessedCount: 1,
		Processed:      map[int]bool{messageID: true},
	}
}

// GetProcessedMessages gets the dedupe data for the specific map and batch
func (d DedupeMap) GetProcessedMessages(mapID string, batchID int) (*DedupeProcessedMessages, bool) {
	processedMessages, ok := d[mapID][batchID]
	return processedMessages, ok
}

// DedupeProcessedMessages holds the processed messages
// for a specific mapper and batch
type DedupeProcessedMessages struct {
	ProcessedCount int
	Processed      map[int]bool
}

// IsBatchComplete checks if the reducer has processed
// the maximum amount of message a batch can have
func (dp *DedupeProcessedMessages) IsBatchComplete() bool {
	if dp.ProcessedCount == MaxItemsPerBatch {
		return true
	}
	return false
}

// IsMessageProcessed returns true if the message has been processed
func (dp *DedupeProcessedMessages) IsMessageProcessed(mesageID int) bool {
	return dp.Processed[mesageID]
}

// UpdateMessageProcessed updates the dedupe map to register the
// given message as registered
func (dp *DedupeProcessedMessages) UpdateMessageProcessed(mesageID int) {
	dp.Processed[mesageID] = true
	dp.ProcessedCount = dp.ProcessedCount + 1
}

// DeletedProcessedMessages deletes the processed messages map
// this should be called if IsBatchComplete() returns true
// and it is used to save memory space
func (dp *DedupeProcessedMessages) DeletedProcessedMessages() {
	dp.Processed = nil
}
