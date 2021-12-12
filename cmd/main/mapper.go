package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	conf "github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/driver"
)

// TODO: create different logic when full object is provided
// this is needed to allow objects to be downloaded concurrently
// if we use ranges, automatically concurrency does not work

var downloader *manager.Downloader
var uploader *manager.Uploader

func init() {
	// TODO: add proper configuration
	cfg, err := conf.InitLocalCfg()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an S3 client using the loaded configuration
	s3Client := s3.NewFromConfig(cfg)

	// Create a S3 downloader and uploader
	downloader = manager.NewDownloader(s3Client)
	uploader = manager.NewUploader(s3Client)
}

func HandleRequest(ctx context.Context, request driver.Mapping) (string, error) {
	fmt.Println("I AM HERE")
	return "HELLO", nil
	// keep a dictionary with the number of batches
	batchMetadata := make(map[string]int64)

	for _, object := range request.Objects {
		// download file
		err := downloadFile(
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
		mapOutput := map[string]int{
			"hello": 2,
			"hi":    3,
		}

		// send output to reducers via queues
		err = emitMap(mapOutput, request.NumQueues, batchMetadata)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// clean up file in /tmp
		// TODO
	}

	// write batch metadata to S3
	err := writeBatchMetadata(
		ctx,
		request.JobBucket,
		fmt.Sprintf("metadata/%s", request.MapID.String()),
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
	// testing()
}

// TODO add json notation
type mapInt struct {
	key   string
	value int
}

func emitMap(output map[string]int, numQueues int64, batchMetadata map[string]int64) error {
	// keep dictionary of batches to allow sending keys in batches
	batches := make(map[string][]mapInt)

	for key, value := range output {
		// get partition queue from key
		partitionQueue := getQueuePartition(key, numQueues)

		// add to batch
		batches[partitionQueue] = append(
			batches[partitionQueue],
			mapInt{
				key:   key,
				value: value,
			},
		)

		// flush batch if it has 10 items
		if len(batches[partitionQueue]) == 10 {
			// send to queue
			sendBatch()

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
		extraValuesForBatch := make([]mapInt, 10-len(valuesInBatch))
		for i := range extraValuesForBatch {
			extraValuesForBatch[i] = mapInt{} // nil value
		}
		batches[key] = append(batches[key], extraValuesForBatch...)

		// send batch
		sendBatch()

		// update batch metadata
		batchMetadata[key] = batchMetadata[key] + int64(1)
	}

	return nil
}

func sendBatch() {
	// TODO: unimplemented
}

func writeBlankFile() {
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
	partitionQueue := int(bi.Int64() % numQueues)

	return strconv.Itoa(partitionQueue)
}

func downloadFile(bucket, key string, initialByte, finalByte int64) error {
	file, err := os.Create(filepath.Join("/tmp", key))
	if err != nil {
		fmt.Println(err)
		return err
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
		return err
	}

	// check that the bytes read match expectation
	if bytesRead != finalByte-initialByte {
		fmt.Println(err)
		return errors.New("File was not read correctly")
	}

	return nil
}
