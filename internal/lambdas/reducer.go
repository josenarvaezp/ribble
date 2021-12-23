package lambdas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/objectstore"
	"github.com/josenarvaezp/displ/internal/queues"
)

// ReducerInput is the input the reducer lambda receives
type ReducerInput struct {
	JobID          uuid.UUID `json:"jobID"`
	QueuePartition int       `json:"queuePartition"`
	NumMappers     int       `json:"numMappers"`
}

// Reducer is an interface that implements ReducerAPI
type Reducer struct {
	JobID     uuid.UUID
	ReducerID uuid.UUID
	// clients
	DownloaderAPI objectstore.ManagerDownloaderAPI
	UploaderAPI   objectstore.ManagerUploaderAPI
	QueuesAPI     queues.QueuesAPI
	// metadata
	Region         string
	AccountID      string
	NumMappers     int
	QueuePartition int
	Local          bool
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
	mapper := &Reducer{
		Region: region,
		Local:  local,
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

	// Create a S3 downloader and uploader
	mapper.DownloaderAPI = manager.NewDownloader(s3Client)
	mapper.UploaderAPI = manager.NewUploader(s3Client)

	// create sqs client
	mapper.QueuesAPI = sqs.NewFromConfig(cfg)

	return mapper, err
}

// WriteReducerOutput writes the output of the reducer to objectstore
func (r *Reducer) WriteReducerOutput(ctx context.Context, outputMap map[string]int) error {
	// encode map to JSON
	p, err := json.Marshal(outputMap)
	if err != nil {
		return err
	}

	// use uploader manager to write file to S3
	jsonContentType := "application/json"
	bucket := r.JobID.String()
	key := fmt.Sprintf("output/%s", r.ReducerID.String())
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
