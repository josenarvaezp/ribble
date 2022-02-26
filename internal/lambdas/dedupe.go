package lambdas

// Dedupe holds data for the write and read dedupe maps. We use two
// dedupe maps to avoid write conflicts when saving the dedupe data
// in the checkpoints while we still read more messages from sqs.
type Dedupe struct {
	WriteMap DedupeMap
	ReadMap  DedupeMap
}

// InitDedupe initializes a dedupe struct
func InitDedupe() *Dedupe {
	return &Dedupe{
		WriteMap: InitDedupeMap(),
		ReadMap:  InitDedupeMap(),
	}
}

// InitDedupeBatch initializes the read map for the current map and batch
func (d *Dedupe) InitDedupeBatch(mapID string, batchID int, messageID int) {
	// if map doesn't exist for current map id init
	if _, ok := d.WriteMap[mapID]; !ok {
		d.WriteMap[mapID] = make(map[int]*DedupeProcessedMessages)
	}

	// update batch map
	d.WriteMap[mapID][batchID] = &DedupeProcessedMessages{
		ProcessedCount: 1,
		Processed:      map[int]bool{messageID: true},
	}
}

// BatchExists checks if the batch exists either in the read or write map
func (d *Dedupe) BatchExists(mapID string, batchID int) bool {
	if _, ok := d.WriteMap[mapID][batchID]; ok {
		return true
	}

	if _, ok := d.ReadMap[mapID][batchID]; ok {
		return true
	}

	return false
}

// IsBatchComplete checks if the reducer has processed
// the maximum amount of message a batch can have
func (d *Dedupe) IsBatchComplete(mapID string, batchID int) bool {
	writeMapProcessedCount := d.WriteMap[mapID][batchID].ProcessedCount

	readMapProcessedCount := 0
	if _, ok := d.ReadMap[mapID]; ok {
		readMapProcessedCount = d.ReadMap[mapID][batchID].ProcessedCount
	}

	if writeMapProcessedCount+readMapProcessedCount == MaxItemsPerBatch {
		return true
	}

	return false
}

// IsMessageProcessed returns true if the message has been processed
func (d *Dedupe) IsMessageProcessed(mapID string, batchID int, mesageID int) bool {
	if _, ok := d.ReadMap[mapID]; ok {
		return d.WriteMap[mapID][batchID].Processed[mesageID] || d.ReadMap[mapID][batchID].Processed[mesageID]
	}

	return d.WriteMap[mapID][batchID].Processed[mesageID]
}

// GetProcessedMessages gets the dedupe data for the specific map and batch.
// It looks for messages in both read and write maps
func (d *Dedupe) GetProcessedMessages(mapID string, batchID int) (DedupeProcessedMessages, bool) {
	processedMessages := DedupeProcessedMessages{}

	processedMessagesWriteMap, okWrite := d.WriteMap[mapID][batchID]
	if okWrite {
		processedMessages.ProcessedCount = processedMessagesWriteMap.ProcessedCount
		processedMessages.Processed = mergeBoolMaps(processedMessagesWriteMap.Processed, processedMessages.Processed)
	}

	processedMessagesReadMap, okRead := d.ReadMap[mapID][batchID]
	if okRead {
		processedMessages.ProcessedCount = processedMessages.ProcessedCount + processedMessagesReadMap.ProcessedCount
		processedMessages.Processed = mergeBoolMaps(processedMessagesReadMap.Processed, processedMessages.Processed)
	}

	return processedMessages, okWrite || okRead
}

// UpdateMessageProcessed updates the dedupe map to register the
// given message as registered
func (d *Dedupe) UpdateMessageProcessed(mapID string, batchID int, mesageID int) {
	d.WriteMap[mapID][batchID].Processed[mesageID] = true
	d.WriteMap[mapID][batchID].ProcessedCount++
}

// DeletedProcessedMessages deletes the processed messages map
// this should be called if IsBatchComplete() returns true
// and it is used to save memory space
func (d *Dedupe) DeletedProcessedMessages(mapID string, batchID int) {
	d.WriteMap[mapID][batchID].Processed = nil
}

// Merge is used to merge the read and write maps into the read map
func (d *Dedupe) Merge() {
	// update read map with write map values
	for mapperID, batchMap := range d.WriteMap {
		for batchID, dedupeMessages := range batchMap {
			d.ReadMap[mapperID][batchID] = dedupeMessages
		}
	}
}

// DedupeMap is a map holding the values of the already processed
// messages. The key to the first map represents the mappers' IDs.
// The second key represents the batchID for the current mapper.
type DedupeMap map[string]map[int]*DedupeProcessedMessages

// InitDedupeMap initializes a dedupe map
func InitDedupeMap() DedupeMap {
	return make(map[string]map[int]*DedupeProcessedMessages)
}

// DedupeProcessedMessages holds the processed messages
// for a specific mapper and batch
type DedupeProcessedMessages struct {
	ProcessedCount int          `json:"processedCount"`
	Processed      map[int]bool `json:"processed,omitempty"`
}

// mergeBoolMaps is a helper function to merge the input map into the output map
func mergeBoolMaps(input, output map[int]bool) map[int]bool {
	for k, v := range input {
		output[k] = v
	}

	return output
}
