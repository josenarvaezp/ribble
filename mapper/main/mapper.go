package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	conf "github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/driver"
)

// TODO: create different logic when full object is provided
// this is needed to allow objects to be downloaded concurrently
// if we use ranges, automatically concurrency does not work

var downloader *manager.Downloader
var uploader *manager.Uploader
var sqsClient *sqs.Client
var region string

type MapperInput struct {
	JobID   uuid.UUID      `json:"jobID"`
	Mapping driver.Mapping `json:"mapping"`
}

func init() {
	// TODO: add proper configuration
	cfg, err := conf.InitLocalLambdaCfg()
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

func HandleRequest(ctx context.Context, request MapperInput) (string, error) {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return "", errors.New("Error getting lambda context")
	}
	accountID := strings.Split(lc.InvokedFunctionArn, ":")[4]

	fmt.Println("Account id")
	fmt.Println(accountID)

	// keep a dictionary with the number of batches
	batchMetadata := make(map[string]int64)

	for _, object := range request.Mapping.Objects {
		// download file
		filename, err := downloadFile(
			object.Bucket,
			object.Key,
			object.InitialByte,
			object.FinalByte,
		)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// user function starts here
		mapOutput := runMapper(*filename, myfunction)

		// send output to reducers via queues
		err = emitMap(ctx, region, accountID, request.JobID, request.Mapping.MapID, mapOutput, request.Mapping.NumQueues, batchMetadata)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// clean up file in /tmp
		err = os.Remove(*filename)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
	}

	// write batch metadata to S3
	err := writeBatchMetadata(
		ctx,
		request.Mapping.JobBucket,
		fmt.Sprintf("metadata/%s", request.Mapping.MapID.String()),
		batchMetadata,
		uploader,
	)
	if err != nil {
		return "", err
	}

	// check if this mapper is the last one and write blank file
	writeBlankFile()

	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}

func runMapper(filename string, userMap func(filename string) map[string]int) map[string]int {
	return userMap(filename)
}

func myfunction(filename string) map[string]int {
	csvFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	output := make(map[string]int)
	for _, line := range csvLines {
		count, err := strconv.Atoi(line[5])
		if err != nil {
			// ignore value
		}
		output[line[1]] = output[line[1]] + count
	}

	return output
}

type MapInt struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func emitMap(
	ctx context.Context,
	region string,
	accountID string,
	jobID uuid.UUID,
	mapID uuid.UUID,
	output map[string]int,
	numQueues int64,
	batchMetadata map[string]int64,
) error {
	// keep dictionary of batches to allow sending keys in batches
	batches := make(map[string][]MapInt)
	batchCount := 0

	for key, value := range output {
		// get partition queue from key
		partitionQueue := getQueuePartition(key, numQueues)

		// add to batch
		batches[partitionQueue] = append(
			batches[partitionQueue],
			MapInt{
				Key:   key,
				Value: value,
			},
		)

		// flush batch if it has 10 items
		if len(batches[partitionQueue]) == 10 {
			// increase number of batches
			batchCount++

			// send batch to queue
			input := &sendBatchInput{
				jobID:          jobID,
				mapID:          mapID,
				partitionQueue: partitionQueue,
				batchID:        batchCount,
				batch:          batches[partitionQueue],
				region:         region,
				accountID:      accountID,
			}
			if err := sendBatch(ctx, input, sqsClient); err != nil {
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
		// increase number of batches
		batchCount++

		// add values until we complete the batch
		// Note that while this is a little more inefficient for the mapper
		// since we could send batches with less values, the reducers logic will
		// be much simpler given that a reducer will only need to know the number of
		// batches that the mapper sent rather that the number of batches and for
		// each batch how many items
		for i := 0; i < 10-len(valuesInBatch); i++ {
			// TODO: this is sending empty values because of the json encoding
			valuesInBatch = append(valuesInBatch, MapInt{}) // append nil value
		}

		// send batch to queue
		input := &sendBatchInput{
			jobID:          jobID,
			mapID:          mapID,
			partitionQueue: key,
			batchID:        batchCount,
			batch:          batches[key],
			region:         region,
			accountID:      accountID,
		}
		if err := sendBatch(ctx, input, sqsClient); err != nil {
			return err
		}

		// update batch metadata
		batchMetadata[key] = batchMetadata[key] + int64(1)
	}

	return nil
}

type sendBatchInput struct {
	jobID          uuid.UUID
	mapID          uuid.UUID
	partitionQueue string
	batchID        int
	batch          []MapInt
	region         string
	accountID      string
}

func sendBatch(ctx context.Context, input *sendBatchInput, sqsClient *sqs.Client) error {
	numberDataType := "Number"
	stringDataType := "String"

	// convert batch to message entries
	messsageEntries := make([]types.SendMessageBatchRequestEntry, len(input.batch))
	for i, message := range input.batch {
		messageID := strconv.Itoa(i) // unique message id within batch
		mapID := input.mapID.String()
		batchID := strconv.Itoa(input.batchID)

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
				"map-id": {
					DataType:    &stringDataType,
					StringValue: &mapID,
				},
				"batch-id": {
					DataType:    &numberDataType,
					StringValue: &batchID,
				},
			},
		}
	}

	queueName := input.partitionQueue
	// queueURL := fmt.Sprintf(
	// 	"https://sqs.%s.amazonaws.com/%s/%s",
	// 	input.region,
	// 	input.accountID,
	// 	queueName,
	// )
	queueURL := fmt.Sprintf(
		"https://localstack:4566/000000000000/%s-%s", // TODO
		input.jobID.String(),
		queueName,
	)
	params := &sqs.SendMessageBatchInput{
		Entries:  messsageEntries,
		QueueUrl: &queueURL,
	}
	output, err := sqsClient.SendMessageBatch(ctx, params)
	if err != nil {
		return err
	}

	if len(output.Failed) != 0 {
		// TODO: retry
		fmt.Println("some messages were not sent")
	}

	return nil
}

func writeBlankFile() {
	// TODO: list is an expensive operation so maybe there is a nother solution
	// using Dynamo
	// list files in /metadata

	// if files == numMappers then this is the last mapper

	// list files in signals/coordinator/ to check if another mapper
	// at the same time has written this blank file

	// if no files then write new file
}

func writeBatchMetadata(ctx context.Context, bucket, key string, batchMetadata map[string]int64, uploader *manager.Uploader) error {
	// encode map to JSON
	p, err := json.Marshal(batchMetadata)
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

func getQueuePartition(key string, numQueues int64) string {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(key))
	hexstr := hex.EncodeToString(h.Sum(nil))
	bi.SetString(hexstr, 16)
	partitionQueue := int(bi.Uint64() % uint64(numQueues))

	return strconv.Itoa(partitionQueue)
}

func downloadFile(bucket, key string, initialByte, finalByte int64) (*string, error) {
	file, err := os.Create(filepath.Join("/tmp", key))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	objectRange := fmt.Sprintf("bytes=%d-%d", initialByte, finalByte)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Range:  aws.String(objectRange),
	}

	bytesRead, err := downloader.Download(context.Background(), file, input)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// check that the bytes read match expectation
	if bytesRead != finalByte-initialByte {
		fmt.Println(err)
		return nil, errors.New("File was not read correctly")
	}

	filename := file.Name()
	return &filename, nil
}
