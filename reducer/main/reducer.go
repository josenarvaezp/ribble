package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	conf "github.com/josenarvaezp/displ/internal/config"
)

// TODO: add to util
type MapInt struct {
	Key      string `json:"key,omitempty"`
	Value    int    `json:"value,omitempty"`
	EmptyVal bool   `json:"empty,omitempty"`
}

var downloader *manager.Downloader
var uploader *manager.Uploader
var sqsClient *sqs.Client
var region string

type ReducerInput struct {
	JobID     uuid.UUID `json:"jobID"`
	JobBucket string    `json:"jobBucket"` // TODO: remove
	QueueName string    `json:"queueName"`
}

func init() {
	// TODO: add proper configuration
	cfg, err := conf.InitLocalCfg()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an S3 client using the loaded configuration
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Create a S3 downloader and uploader
	downloader = manager.NewDownloader(s3Client)
	uploader = manager.NewUploader(s3Client)

	// create sqs client
	sqsClient = sqs.NewFromConfig(cfg)

	// get region from env var
	region = os.Getenv("AWS_REGION")
}

type dedupeProcessed struct {
	processedCount int
	processed      map[int]bool
}

func HandleRequest(ctx context.Context, request ReducerInput) (string, error) {
	reducerID := uuid.New()
	queueURLOutput, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &request.QueueName,
	})

	totalBatchesToProcess := getNumberOfBatchesToProcess(request.QueueName)
	totalProcessedBatches := 0
	outputMap := make(map[string]int)

	recieveMessageParams := &sqs.ReceiveMessageInput{
		QueueUrl:              queueURLOutput.QueueUrl,
		MaxNumberOfMessages:   10,
		MessageAttributeNames: []string{"map-id", "batch-id", "message-id"},
		// TODO add support for long polling using waitTime,
		// given that by the time we process results they are all
		// in the queues we shouldn't need long polling
	}

	// dedupe map represents a map of all mappers that write to the queue
	// and each map holds a map of the batches that the reducer has processed
	dedupeMap := make(map[string]map[int]*dedupeProcessed)

	// recieve messages until we are done processing all queue
	for true {
		output, err := sqsClient.ReceiveMessage(ctx, recieveMessageParams)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		for _, message := range output.Messages {
			// unmarshall body
			var res MapInt
			body := []byte(*message.Body)
			err = json.Unmarshal(body, &res)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			// get message attributes
			currentMapID := message.MessageAttributes["map-id"].StringValue
			currentBatchID, err := strconv.Atoi(*message.MessageAttributes["batch-id"].StringValue)
			if err != nil {
				fmt.Println(err)
				return "", err
			}
			currentMessageID, err := strconv.Atoi(*message.MessageAttributes["message-id"].StringValue)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			// check if message has already been processed
			dedupeVal, ok := dedupeMap[*currentMapID][currentBatchID]
			if ok {
				// check if the count of processed values is less than 10
				// as only 10 values are allowed in each batch
				if dedupeVal.processedCount == 10 {
					// ignore as it is a duplicated message
					continue
				}

				// check if message id is present in dedupe map
				if dedupeVal.processed[currentMessageID] {
					// ignore as it is a duplicated message
					continue
				}

				// process message
				currentKey := res.Key
				currentValue := res.Value

				// only process value if it is not empty
				// emty values are sent to keep the same number of events per batch
				if res.EmptyVal != true {
					outputMap[currentKey] = outputMap[currentKey] + currentValue
				}

				// add message to dedupe map
				dedupeMap[*currentMapID][currentBatchID].processed[currentMessageID] = true
				dedupeMap[*currentMapID][currentBatchID].processedCount = dedupeVal.processedCount + 1

				// check if we are done processing batch from map
				if dedupeMap[*currentMapID][currentBatchID].processedCount == 10 {
					totalProcessedBatches++
					// delete processed map
					// TODO: delete map
					// delete(dedupeMap[*currentMapID][currentBatchID].processed)
				}

			} else {
				// no message from current batch has been processed
				// process message
				currentKey := res.Key
				currentValue := res.Value

				// only process value if it is not empty
				// emty values are sent to keep the same number of events per batch
				if res.EmptyVal != true {
					outputMap[currentKey] = outputMap[currentKey] + currentValue
				}

				// init dedupe data for batch
				dedupeMap[*currentMapID] = make(map[int]*dedupeProcessed)
				dedupeMap[*currentMapID][currentBatchID] = &dedupeProcessed{
					processedCount: 1,
					processed:      map[int]bool{currentMessageID: true},
				}
			}
		}

		// check if we are done processing values
		if totalProcessedBatches == totalBatchesToProcess {
			break
		}
	}

	outputKey := fmt.Sprintf("output/%s", reducerID.String())
	err = writeReducerOutput(ctx, request.JobBucket, outputKey, outputMap, uploader)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return "", nil
}

func main() {
	ctx := context.Background()
	request := ReducerInput{
		JobID:     uuid.MustParse("1469d3b8-d133-4036-8944-01ff6518ec25"),
		JobBucket: "jobbucket",
		QueueName: "1469d3b8-d133-4036-8944-01ff6518ec25-4",
	}
	HandleRequest(ctx, request)
	// lambda.Start(HandleRequest)
}

func getQueueMetadataFromS3(ctx context.Context, bucket string, Key string, s3Client *s3.Client) (map[string]int64, error) {
	contentType := "application/json"
	params := &s3.GetObjectInput{
		Bucket:              &bucket, // this is the job bucket
		Key:                 &Key,    // this is the queue name
		ResponseContentType: &contentType,
	}
	output, err := s3Client.GetObject(ctx, params)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	var batchMetadata map[string]int64
	err = json.Unmarshal(body, batchMetadata)
	if err != nil {
		return nil, err
	}

	return batchMetadata, nil
}

func getQueueMetadataFromQueue(ctx context.Context, queueURL string, sqsClient *sqs.Client) (map[string]int64, error) {
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL, // name of the fifo queue TODO probs jobid.Fifo
		AttributeNames:      []sqsTypes.QueueAttributeName{},
		MaxNumberOfMessages: 10,

		// TODO: https://rcdexta.medium.com/processing-high-volume-of-unique-messages-exactly-once-while-preserving-order-in-a-queue-d8d6184ded01

	}
	sqsClient.ReceiveMessage(ctx, params)
	return nil, nil
}

func getNumberOfBatchesToProcess(queueName string) int {
	// TODO
	return 2
}

// TODO: use this function from utils
func writeReducerOutput(ctx context.Context, bucket, key string, outputMap map[string]int, uploader *manager.Uploader) error {
	// encode map to JSON
	p, err := json.Marshal(outputMap)
	if err != nil {
		return err
	}

	// use uploader manager to write file to S3
	jsonContentType := "application/json"
	input := &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          bytes.NewReader(p),
		ContentType:   &jsonContentType,
		ContentLength: int64(len(p)),
	}
	_, err = uploader.Upload(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
