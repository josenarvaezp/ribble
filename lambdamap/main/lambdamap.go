package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/josenarvaezp/displ/internal/lambdas"
)

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
		mapOutput := runMapper(*filename, WordCount)

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

func WordCount(filename string) map[string]int {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	output := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		for _, word := range words {
			output[word] = output[word] + 1
		}
	}

	return output
}
