package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/driver"
	intMapper "github.com/josenarvaezp/displ/internal/mapper"
)

// TODO: create different logic when full object is provided
// this is needed to allow objects to be downloaded concurrently
// if we use ranges, automatically concurrency does not work

// TODO: rename this package to something else
var mapper *intMapper.Mapper

type MapperInput struct {
	JobID   uuid.UUID      `json:"jobID"`
	Mapping driver.Mapping `json:"mapping"`
}

func init() {
	var err error
	mapper, err = intMapper.NewMapper(true)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func HandleRequest(ctx context.Context, request MapperInput) (string, error) {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return "", errors.New("Error getting lambda context")
	}
	mapper.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	mapper.JobID = request.JobID
	mapper.MapID = request.Mapping.MapID
	mapper.NumQueues = request.Mapping.NumQueues

	// keep a dictionary with the number of batches per queue
	batchMetadata := make(map[string]int64)

	for _, object := range request.Mapping.Objects {
		// download file
		filename, err := mapper.DownloadFile(object)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// user function starts here
		mapOutput := runMapper(*filename, myfunction)

		// send output to reducers via queues
		err = mapper.EmitMap(ctx, mapOutput, batchMetadata)
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

	// send batch metadata to S3
	err := mapper.WriteBatchMetadata(
		ctx,
		request.Mapping.JobBucket,
		fmt.Sprintf("metadata/%s", request.Mapping.MapID.String()),
		batchMetadata,
	)
	if err != nil {
		return "", err
	}

	// check if this mapper is the last one and write blank file
	mapper.WriteBlankFile()

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
