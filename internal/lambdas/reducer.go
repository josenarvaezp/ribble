package lambdas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/queues"
)

const (
	// constants used for doing the checkpoint mechanism
	MaxMessagesWithoutCheckpoint        = 100000
	MaxMessagesBeforeCheckpointComplete = 15000
)

// ReducerInput is the input the reducer lambda receives
type ReducerInput struct {
	JobID          uuid.UUID `json:"jobID"`
	ReducerID      uuid.UUID `json:"reducerID"`
	QueuePartition int       `json:"queuePartition"`
	NumMappers     int       `json:"numMappers"`
}

// Reducer is an interface that implements ReducerAPI
type Reducer struct {
	JobID     uuid.UUID
	ReducerID uuid.UUID
	// clients
	ObjectStoreAPI objectstore.ObjectStoreAPI
	DownloaderAPI  objectstore.ManagerDownloaderAPI
	UploaderAPI    objectstore.ManagerUploaderAPI
	QueuesAPI      queues.QueuesAPI
	// metadata
	Region         string
	AccountID      string
	NumMappers     int
	QueuePartition int
	Local          bool
	OutputMap      map[string]int
	Dedupe         *Dedupe
	mu             sync.Mutex
}

// NewReducer initializes a new reducer with its required clients
func NewReducer(
	local bool,
) (*Reducer, error) {
	var cfg aws.Config
	var err error

	// get region from env var
	region := os.Getenv("AWS_REGION")

	// init mapper
	reducer := &Reducer{
		Region:    region,
		Local:     local,
		OutputMap: make(map[string]int),
		Dedupe:    InitDedupe(),
	}

	// create config
	if local {
		cfg, err = config.InitLocalCfg()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.InitCfg(region)
		if err != nil {
			return nil, err
		}
	}

	// Create an S3 client using the loaded configuration
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	reducer.ObjectStoreAPI = s3Client

	// Create a S3 downloader and uploader
	reducer.DownloaderAPI = manager.NewDownloader(s3Client)
	reducer.UploaderAPI = manager.NewUploader(s3Client)

	// create sqs client
	reducer.QueuesAPI = sqs.NewFromConfig(cfg)

	return reducer, err
}

// WriteReducerOutput writes the output of the reducer to objectstore
func (r *Reducer) WriteReducerOutput(ctx context.Context, outputMap map[string]int, key string) error {
	// encode map to JSON
	p, err := json.Marshal(outputMap)
	if err != nil {
		return err
	}

	// use uploader manager to write file to S3
	jsonContentType := "application/json"
	bucket := r.JobID.String()
	input := &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          bytes.NewReader(p),
		ContentType:   &jsonContentType,
		ContentLength: int64(len(p)),
	}
	_, err = r.UploaderAPI.Upload(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

// GetNumberOfBatchesToProcess gets the number of batches a the reducer needs to process
// based on the metadata available in the metadata queue for that reducer
func (r *Reducer) GetNumberOfBatchesToProcess(ctx context.Context) (*int, error) {
	// holds number of messages to process
	totalNumOfMessagesToProcess := 0

	// receive message params
	queueName := fmt.Sprintf("%s-%d-meta", r.JobID.String(), r.QueuePartition)
	queueURL := GetQueueURL(queueName, r.Region, r.AccountID, r.Local)
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: MaxItemsPerBatch,
	}

	// dedupeMap is used to check if we have processed a message already
	dedupeMap := make(map[string]bool)
	mappersProccessedCount := 0

	// get metadata until we have metadata from each mapper
	for mappersProccessedCount != r.NumMappers {
		// haven't recived all metadata from all mappers
		output, err := r.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, message := range output.Messages {

			// unmarshal metadata message
			var res QueueMetadata
			body := []byte(*message.Body)
			err = json.Unmarshal(body, &res)
			if err != nil {
				return nil, err
			}

			// add to totalNumOfMessagesToProcess if we have not
			// processed the current message already
			if _, ok := dedupeMap[res.MapID]; !ok {
				dedupeMap[res.MapID] = true
				totalNumOfMessagesToProcess = totalNumOfMessagesToProcess + res.NumBatches
				mappersProccessedCount++
			}
		}
	}

	return &totalNumOfMessagesToProcess, nil
}

// SaveIntermediateOutput saves the intermediate output into an S3 object
func (r *Reducer) SaveIntermediateOutput(
	ctx context.Context,
	intermediateMap map[string]int,
	currentCheckpoint int,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	// save intermediate output map
	key := fmt.Sprintf("checkpoints/%s/%d-intermediate", r.ReducerID.String(), currentCheckpoint)
	if err := r.WriteReducerOutput(ctx, intermediateMap, key); err != nil {
		return err
	}

	return nil
}

// SaveIntermediateDedupe saves the intermediate dedupe data to an S3 file
func (r *Reducer) SaveIntermediateDedupe(ctx context.Context, currentCheckpoint int, wg *sync.WaitGroup) error {
	defer wg.Done()

	// encode map to JSON
	p, err := json.Marshal(r.Dedupe.WriteMap)
	if err != nil {
		return err
	}

	// use uploader manager to write file to S3
	jsonContentType := "application/json"
	bucket := r.JobID.String()
	key := fmt.Sprintf("checkpoints/%s/%d-dedupe", r.ReducerID.String(), currentCheckpoint)
	inputParams := &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          bytes.NewReader(p),
		ContentType:   &jsonContentType,
		ContentLength: int64(len(p)),
	}
	_, err = r.UploaderAPI.Upload(ctx, inputParams)
	if err != nil {
		return err
	}

	return nil
}

// UpdateOutputMap merges the outputMap with the intermediate map
func (r *Reducer) UpdateOutputMap(intermediateMap map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()

	// update output map values
	for k, v := range intermediateMap {
		r.OutputMap[k] = r.OutputMap[k] + v
	}
}

// DeleteIntermediateMessagesFromQueue deletes the read messages from sqs
func (r *Reducer) DeleteIntermediateMessagesFromQueue(
	ctx context.Context,
	queueURL string,
	deleteEntries []sqsTypes.DeleteMessageBatchRequestEntry,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	params := &sqs.DeleteMessageBatchInput{
		QueueUrl: &queueURL,
	}

	firstMessageToDelete := 0
	lastMessageToDelete := 10

	// we can only delete 10 per call so we need to loop through all delete requests
	for lastMessageToDelete <= len(deleteEntries) {
		messagesToDelete := deleteEntries[firstMessageToDelete:lastMessageToDelete]
		params.Entries = messagesToDelete
		_, err := r.QueuesAPI.DeleteMessageBatch(ctx, params)
		if err != nil {
			return err
		}

		// TODO check failed results and add error message - we probs don't want to stop execution

		// update messages to delete indexes
		firstMessageToDelete = lastMessageToDelete

		if len(deleteEntries) < lastMessageToDelete+10 {
			// we don't have 10 values to delete
			lastMessageToDelete = lastMessageToDelete + (len(deleteEntries) - lastMessageToDelete)
		} else {
			lastMessageToDelete = lastMessageToDelete + 10
		}
	}

	return nil
}

// CheckpointData is used to hold checkpoint objects from s3
type CheckpointData struct {
	LastCheckpoint         int
	DedupeData             []objectstore.Object
	IntermediateOutputData []objectstore.Object
}

// GetCheckpointData gets all the checkpoint objects for the reducer and
// returns a CheckpointData struct which holds data about the intermediate and
// dedupe objects
func (r *Reducer) GetCheckpointData(ctx context.Context) (*CheckpointData, error) {
	// used to split objects
	checkpointData := &CheckpointData{
		LastCheckpoint:         0,
		DedupeData:             []objectstore.Object{},
		IntermediateOutputData: []objectstore.Object{},
	}

	// used for pagination in the list objects call
	var continuationToken *string

	// indifcates if there are more objects to be listed
	moreObjects := true

	bucket := r.JobID.String()
	prefixKey := fmt.Sprintf("checkpoints/%s", r.ReducerID.String())

	for moreObjects {
		params := &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			MaxKeys:           1000,
			ContinuationToken: continuationToken,
			Prefix:            &prefixKey,
		}
		listObjectsOuput, err := r.ObjectStoreAPI.ListObjectsV2(ctx, params)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		// convert object
		objects := objectstore.S3ObjectsToObjects(bucket, listObjectsOuput.Contents)

		// divide objects into dedupe and intermediate data
		for _, object := range objects {
			if strings.Contains(object.Key, "dedupe") {
				checkpointData.DedupeData = append(checkpointData.DedupeData, object)
				checkpointData.LastCheckpoint++
			} else if strings.Contains(object.Key, "intermediate") {
				checkpointData.IntermediateOutputData = append(checkpointData.IntermediateOutputData, object)
			}
		}

		// update pagination token
		continuationToken = listObjectsOuput.NextContinuationToken

		// check if there are more objects remaining
		moreObjects = listObjectsOuput.IsTruncated
	}

	return checkpointData, nil
}

// GetOutputMap updates the output map with the data from the intermediate checkpoints.
// For each intermediate checkpoint it merges the data with the output map concurrently
func (r *Reducer) GetOutputMap(ctx context.Context, intermediateData []objectstore.Object, wg *sync.WaitGroup) error {

	// loop through intermediate results
	for _, intermediateOutputObject := range intermediateData {
		wg.Add(1)
		go r.updateOutputWithIntermediateObject(ctx, intermediateOutputObject, wg)
	}

	return nil
}

// updateOutputWithIntermediateObject is a helper function to merge the outputData concurrently.
// It downloads a intermediate object and merges the data to the outputMap with a mutex
// so that the data is updated consistently accross all go routines
func (r *Reducer) updateOutputWithIntermediateObject(
	ctx context.Context,
	intermediateOutputObject objectstore.Object,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	params := &s3.GetObjectInput{
		Bucket: &intermediateOutputObject.Bucket,
		Key:    &intermediateOutputObject.Key,
	}
	buf := manager.NewWriteAtBuffer([]byte{})
	_, err := r.DownloaderAPI.Download(ctx, buf, params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// unmarshal result
	var res map[string]int
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return err
	}

	// update output map values
	// use mutex to get consistent result
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range res {
		r.OutputMap[k] = r.OutputMap[k] + v
	}

	return nil
}

// GetDedupe updates the dedupe map with the data from the checkpoints.
// For each intermediate checkpoint it merges the data with the dedupe map concurrently
func (r *Reducer) GetDedupe(ctx context.Context, dedupeData []objectstore.Object, wg *sync.WaitGroup) error {
	// loop through dedupe results
	for _, dedupeObject := range dedupeData {
		wg.Add(1)
		go r.updateDedupeReaderWithDedupeObject(ctx, dedupeObject, wg)
	}

	return nil
}

// updateDedupeReaderWithDedupeObject is a helper function to merge the dedupe map concurrently.
// It downloads a dedupe object and merges the data to the dedupe map with a mutex
// so that the data is updated consistently accross all go routines
func (r *Reducer) updateDedupeReaderWithDedupeObject(ctx context.Context, dedupeObject objectstore.Object, wg *sync.WaitGroup) error {
	defer wg.Done()

	params := &s3.GetObjectInput{
		Bucket: &dedupeObject.Bucket,
		Key:    &dedupeObject.Key,
	}
	buf := manager.NewWriteAtBuffer([]byte{})
	_, err := r.DownloaderAPI.Download(ctx, buf, params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// unmarshal result
	var res DedupeMap
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return err
	}

	// update output map values
	// use mutex to get consistent result
	r.mu.Lock()
	defer r.mu.Unlock()

	// update map
	for mapperID, batchMap := range res {
		for batchID, dedupeMessages := range batchMap {
			if _, ok := r.Dedupe.ReadMap[mapperID]; !ok {
				r.Dedupe.ReadMap[mapperID] = make(map[int]*DedupeProcessedMessages)
			}
			r.Dedupe.ReadMap[mapperID][batchID] = dedupeMessages
		}
	}

	return nil
}