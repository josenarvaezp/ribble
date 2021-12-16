package reducer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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

// Reducer is an interface that implements ReducerAPI
type Reducer struct {
	JobID     uuid.UUID
	ReducerID uuid.UUID
	// clients
	DownloaderAPI objectstore.ManagerDownloaderAPI
	UploaderAPI   objectstore.ManagerUploaderAPI
	QueuesAPI     queues.QueuesAPI
	// metadata
	Region    string
	AccountID string
	Local     bool
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
		cfg, err = config.InitLocalLambdaCfg()
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

func (r *Reducer) GetNumberOfBatchesToProcess(queueName string) int {
	// TODO: read metadata for queue, either from s3, sqs or Dynamo
	return 2
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
