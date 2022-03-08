package lambdas

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/queues"
	"github.com/josenarvaezp/displ/pkg/aggregators"
)

// Mapping represents the collection of objects that are used as input
// for the Mapping stage of the framework. Each mapper recieves an
// input which may contain one or multiple objects, depeding on their size.
type Mapping struct {
	MapID   uuid.UUID                 `json:"id"`
	Objects []objectstore.ObjectRange `json:"rangeObjects"`
	Size    int64                     `json:"size,string"`
}

// NewMapping initialises the M with an id and size 0
func NewMapping() *Mapping {
	return &Mapping{
		MapID: uuid.New(),
		Size:  0,
	}
}

// MapperInput is the input the mapper lambda receives
type MapperInput struct {
	JobID     uuid.UUID `json:"jobID"`
	Mapping   Mapping   `json:"mapping"`
	NumQueues int64     `json:"queues,string"`
}

// MapperAPI is an interface deining the functions available to the mapper
type MapperAPI interface {
	DownloadFile(object objectstore.ObjectRange) (*string, error)
	EmitMap(ctx context.Context, outputMap map[string]int, batchMetadata map[int]int64) error
	WriteBatchMetadata(ctx context.Context, bucket, key string, batchMetadata map[int]int64) error
	SendFinishedEvent(ctx context.Context) error
}

// Mapper is an interface that implements MapperAPI
type Mapper struct {
	JobID uuid.UUID
	MapID uuid.UUID
	// clients
	DownloaderAPI objectstore.ManagerDownloaderAPI
	UploaderAPI   objectstore.ManagerUploaderAPI
	QueuesAPI     queues.QueuesAPI
	// metadata
	Region    string
	AccountID string
	NumQueues int64
	local     bool
}

// NewMapper initializes a new mapper with its required clients
func NewMapper(
	local bool,
) (*Mapper, error) {
	var cfg *aws.Config
	var err error

	// get region from env var
	region := os.Getenv("AWS_REGION")

	// init mapper
	mapper := &Mapper{
		Region: region,
		local:  local,
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
	s3Client := s3.NewFromConfig(*cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Create a S3 downloader and uploader
	mapper.DownloaderAPI = manager.NewDownloader(s3Client)
	mapper.UploaderAPI = manager.NewUploader(s3Client)

	// create sqs client
	mapper.QueuesAPI = sqs.NewFromConfig(*cfg)

	return mapper, err
}

// UpdateMapperWithRequest updates the mapper struct with the information
// gathered from the context and request
func (m *Mapper) UpdateMapperWithRequest(ctx context.Context, request MapperInput) error {
	// get data from context
	if m.local {
		m.AccountID = "000000000000"
	} else {
		lc, ok := lambdacontext.FromContext(ctx)
		if !ok {
			return errors.New("Error getting lambda context")
		}
		m.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	}

	m.JobID = request.JobID
	m.MapID = request.Mapping.MapID
	m.NumQueues = request.NumQueues

	return nil
}

// DownloadFile downloads a file from the object store into the local filesystem
func (m *Mapper) DownloadFile(object objectstore.ObjectRange) (*string, error) {
	// create temporary file to store object
	file, err := os.Create(filepath.Join("/tmp", object.Key))
	if err != nil {
		return nil, err
	}

	// download object accordint to range
	objectRange := fmt.Sprintf("bytes=%d-%d", object.InitialByte, object.FinalByte)
	input := &s3.GetObjectInput{
		Bucket: aws.String(object.Bucket),
		Key:    aws.String(object.Key),
		Range:  aws.String(objectRange),
	}
	_, err = m.DownloaderAPI.Download(context.Background(), file, input)
	if err != nil {
		return nil, err
	}

	filename := file.Name()
	return &filename, nil
}

// EmitMapSum sends the output map in batches to the queues
func (m *Mapper) EmitMap(
	ctx context.Context,
	outputMap aggregators.MapAggregator,
	batchMetadata map[int]int64,
) error {
	// keep dictionary of batches to allow sending keys in batches
	batches := make(map[int][]aggregators.ReduceMessage)

	// iterate through the output map and send values in batches
	for key, value := range outputMap {
		// get partition queue from key
		partitionQueue := m.getQueuePartition(key)

		aggregatorType := GetAggregatorType(value)

		// add value to batch
		mapMessage := aggregators.ReduceMessage{
			Key:  key,
			Type: int64(aggregatorType),
		}

		if aggregatorType == AvgAggregator {
			castAvg := value.(*aggregators.Avg)
			mapMessage.Value = castAvg.GetSum()
			mapMessage.Count = castAvg.GetCount()
		} else {
			mapMessage.Value = value.ToNum()
		}

		batches[partitionQueue] = append(
			batches[partitionQueue],
			mapMessage,
		)

		// flush batch if it has maximum items
		if len(batches[partitionQueue]) == MaxItemsPerBatch {
			// send batch to queue
			if err := m.sendBatch(
				ctx,
				partitionQueue,
				int(batchMetadata[partitionQueue]+int64(1)),
				batches[partitionQueue],
			); err != nil {
				return err
			}

			// update batch metadata
			batchMetadata[partitionQueue] = batchMetadata[partitionQueue] + int64(1)

			// delete batch from map
			delete(batches, partitionQueue)
		}
	}

	// flush all remaining batches that don't have 10 values
	for key, valuesInBatch := range batches {
		// add values until we complete the batch
		// Note that while this is a little more inefficient for the mapper
		// since we could send batches with less values, the reducers logic will
		// be much simpler given that a reducer will only need to know the number of
		// batches that the mapper sent rather that the number of batches and for
		// each batch how many items
		valuesToAppend := make([]aggregators.ReduceMessage, MaxItemsPerBatch-len(valuesInBatch))
		for i := 0; i < len(valuesToAppend); i++ {
			valuesToAppend[i] = aggregators.ReduceMessage{
				EmptyVal: true,
			}
		}

		// append values to values in batch
		valuesInBatch = append(valuesInBatch, valuesToAppend...)

		// send batch to queue
		if err := m.sendBatch(
			ctx,
			key,
			int(batchMetadata[key]+int64(1)),
			valuesInBatch,
		); err != nil {
			return err
		}

		// update batch metadata
		batchMetadata[key] = batchMetadata[key] + int64(1)
	}

	return nil
}

// sendBatch sends the specified batch to the specified queue
func (m *Mapper) sendBatch(ctx context.Context, partitionQueue int, batchID int, batch []aggregators.ReduceMessage) error {
	// convert batch to message entries
	messsageEntries := make([]types.SendMessageBatchRequestEntry, len(batch))
	for i, message := range batch {
		messageID := strconv.Itoa(i) // unique message id within batch
		batchID := strconv.Itoa(batchID)
		mapID := m.MapID.String()

		// encode map input into JSON
		p, err := json.Marshal(message)
		if err != nil {
			return err
		}
		messageJSONString := string(p)

		messsageEntries[i] = types.SendMessageBatchRequestEntry{
			Id:          &messageID,
			MessageBody: &messageJSONString,
			MessageAttributes: map[string]types.MessageAttributeValue{
				MapIDAttribute: {
					DataType:    &stringDataType,
					StringValue: &mapID,
				},
				BatchIDAttribute: {
					DataType:    &numberDataType,
					StringValue: &batchID,
				},
				MessageIDAttribute: {
					DataType:    &numberDataType,
					StringValue: &messageID,
				},
			},
		}
	}

	queueName := fmt.Sprintf("%s-%d", m.JobID.String(), partitionQueue)
	queueURL := GetQueueURL(queueName, m.Region, m.AccountID, m.local)
	params := &sqs.SendMessageBatchInput{
		Entries:  messsageEntries,
		QueueUrl: &queueURL,
	}
	output, err := m.QueuesAPI.SendMessageBatch(ctx, params)
	if err != nil {
		return err
	}

	if len(output.Failed) != 0 {
		// TODO: retry
		fmt.Println("some messages were not sent")
	}

	return nil
}

// getQueuePartition is a helpder function for the mapper that
// gets the queue partition of a key given its md5 hash
func (m *Mapper) getQueuePartition(key string) int {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(key))
	hexstr := hex.EncodeToString(h.Sum(nil))
	bi.SetString(hexstr, 16)
	partitionQueue := int(bi.Uint64() % uint64(m.NumQueues))

	return partitionQueue
}

// GetRandomQueuePartition generates a random number
// by using a generated uuid as the seed for the random
// number generator
func (m *Mapper) GetRandomQueuePartition() int {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(uuid.New().String()))
	hexstr := hex.EncodeToString(h.Sum(nil))
	bi.SetString(hexstr, 16)

	newSource := rand.NewSource(int64(bi.Uint64()))
	randomWithSeed := rand.New(newSource)
	return randomWithSeed.Intn(int(m.NumQueues))
}

// EmitRandom sends the data (a single value) produced by a mapper
// to random partition queues. This is used when the distribution of
// keys is not good to create good load balancing.
func (m *Mapper) EmitRandom(ctx context.Context, outputMap aggregators.MapAggregator, messageMetadata map[int]int64) error {
	for key, value := range outputMap {
		messageID := uuid.New().String()

		// get partition queue from key
		partitionQueue := m.GetRandomQueuePartition()

		aggregatorType := GetAggregatorType(value)

		// add value to batch
		mapMessage := aggregators.ReduceMessage{
			Key:  key,
			Type: int64(aggregatorType),
		}

		if aggregatorType == AvgAggregator {
			castAvg := value.(*aggregators.Avg)
			mapMessage.Value = castAvg.GetSum()
			mapMessage.Count = castAvg.GetCount()
		} else {
			mapMessage.Value = value.ToNum()
		}

		// encode map input into JSON
		p, err := json.Marshal(mapMessage)
		if err != nil {
			return err
		}
		messageJSONString := string(p)

		queueName := fmt.Sprintf("%s-%d", m.JobID.String(), partitionQueue)
		queueURL := GetQueueURL(queueName, m.Region, m.AccountID, m.local)
		params := &sqs.SendMessageInput{
			MessageBody: &messageJSONString,
			MessageAttributes: map[string]types.MessageAttributeValue{
				MessageIDAttribute: {
					DataType:    &stringDataType,
					StringValue: &messageID,
				},
			},
			QueueUrl: &queueURL,
		}
		_, err = m.QueuesAPI.SendMessage(ctx, params)
		if err != nil {
			return err
		}

		// update message metadata
		messageMetadata[partitionQueue] = messageMetadata[partitionQueue] + int64(1)
	}

	return nil
}

// SendMetadata sends the metadata to the metadata queues for mappers
// that work on single values
func (m *Mapper) SendMetadata(ctx context.Context, partition int) error {

	// loop through the queues
	for i := 0; i < int(m.NumQueues); i++ {
		// send params
		queueName := fmt.Sprintf("%s-%d-meta", m.JobID.String(), i)
		queueURL := GetQueueURL(queueName, m.Region, m.AccountID, m.local)
		params := &sqs.SendMessageInput{
			QueueUrl: &queueURL,
		}

		meta := &QueueMetadataSingleValue{
			MapID: m.MapID.String(),
			Sent:  false,
		}

		// mapper sent value to this partition
		if partition == i {
			meta.Sent = true
		}

		// encode metadata into JSON
		p, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		metaJSONString := string(p)

		// add metadata to body
		params.MessageBody = &metaJSONString

		_, err = m.QueuesAPI.SendMessage(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendFinishedEvent sends an event to the mappers-done queue to indicate
// that the current mappers has finished processing
func (m *Mapper) SendFinishedEvent(ctx context.Context) error {
	queueName := fmt.Sprintf("%s-%s", m.JobID.String(), "mappers-done")
	queueURL := GetQueueURL(queueName, m.Region, m.AccountID, m.local)
	curentMapID := m.MapID.String()
	params := &sqs.SendMessageInput{
		MessageBody: &curentMapID,
		QueueUrl:    &queueURL,
	}
	_, err := m.QueuesAPI.SendMessage(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// QueueMetadata is used to send events to the metadata queues
// about how many batches a map processed
type QueueMetadata struct {
	MapID      string `json:"mapID"`
	NumBatches int    `json:"numBatches"`
}

// QueueMetadataSingleValue is used to send events to the metadata queues
// and it indicates if a message was sent
type QueueMetadataSingleValue struct {
	MapID string `json:"mapID"`
	Sent  bool   `json:"sent"`
}

// SendBatchMetadata sends the number of batches the current mapper sent to each of the queues
// this is used so that the reducers know how many events they should process before
// writing out the output
func (m *Mapper) SendBatchMetadata(ctx context.Context, batchMetadata map[int]int64) error {
	meta := &QueueMetadata{
		MapID: m.MapID.String(),
	}

	// loop through the queues
	for i := 0; i < int(m.NumQueues); i++ {
		// send params
		queueName := fmt.Sprintf("%s-%d-meta", m.JobID.String(), i)
		queueURL := GetQueueURL(queueName, m.Region, m.AccountID, m.local)
		params := &sqs.SendMessageInput{
			QueueUrl: &queueURL,
		}

		// add number of batches
		if numOfBatches, ok := batchMetadata[i]; ok {
			meta.NumBatches = int(numOfBatches)
		} else {
			// no message was sent from this mapper to the current queue
			meta.NumBatches = 0
		}

		// encode metadata into JSON
		p, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		metaJSONString := string(p)

		// add metadata to body
		params.MessageBody = &metaJSONString

		_, err = m.QueuesAPI.SendMessage(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunMapAggregator(filename string, userMap func(filename string) aggregators.MapAggregator) aggregators.MapAggregator {
	return userMap(filename)
}
