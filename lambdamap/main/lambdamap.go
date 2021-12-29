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
	"github.com/josenarvaezp/displ/internal/lambdas"
)

// TODO: create different logic when full object is provided
// this is needed to allow objects to be downloaded concurrently
// if we use ranges, automatically concurrency does not work

var m *lambdas.Mapper

func init() {
	var err error
	m, err = lambdas.NewMapper(true)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.MapperInput) (string, error) {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return "", errors.New("Error getting lambda context")
	}
	m.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	m.JobID = request.JobID
	m.MapID = request.Mapping.MapID
	m.NumQueues = request.Mapping.NumQueues

	// keep a dictionary with the number of batches per queue
	batchMetadata := make(map[int]int64)

	for _, object := range request.Mapping.Objects {
		// download file
		filename, err := m.DownloadFile(object)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// user function starts here
		mapOutput := runMapper(*filename, myfunction)

		// send output to reducers via queues
		err = m.EmitMap(ctx, mapOutput, batchMetadata)
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

	// send batch metadata to sqs
	if err := m.SendBatchMetadata(ctx, batchMetadata); err != nil {
		return "", err
	}

	// send event to queue indicating this mapper has completed
	if err := m.SendFinishedEvent(ctx); err != nil {
		return "", err
	}

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