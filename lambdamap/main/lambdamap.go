package main

import (
	"bufio"
	"context"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	log "github.com/sirupsen/logrus"

	"github.com/josenarvaezp/displ/internal/lambdas"
)

var m *lambdas.Mapper

func init() {
	// set logger
	log.SetLevel(log.ErrorLevel)

	var err error
	m, err = lambdas.NewMapper(true)
	if err != nil {
		log.WithError(err).Fatal("Error starting mapper")
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.MapperInput) error {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return errors.New("Error getting lambda context")
	}
	m.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	m.JobID = request.JobID
	m.MapID = request.Mapping.MapID
	m.NumQueues = request.Mapping.NumQueues

	mapperLogger := log.WithFields(log.Fields{
		"Job ID": m.JobID.String(),
	})

	// keep a dictionary with the number of batches per queue
	batchMetadata := make(map[int]int64)

	for _, object := range request.Mapping.Objects {
		// download file
		filename, err := m.DownloadFile(object)
		if err != nil {
			mapperLogger.
				WithFields(log.Fields{
					"Bucket": object.Bucket,
					"Object": object.Key,
				}).
				WithError(err).
				Error("Error downloading file")
			return err
		}

		// user function starts here
		mapOutput := runMapper(*filename, WordCount)

		// send output to reducers via queues
		err = m.EmitMap(ctx, mapOutput, batchMetadata)
		if err != nil {
			mapperLogger.
				WithFields(log.Fields{
					"Bucket": object.Bucket,
					"Object": object.Key,
				}).
				WithError(err).
				Error("Error sending map output to reducers")
			return err
		}

		// clean up file in /tmp
		err = os.Remove(*filename)
		if err != nil {
			mapperLogger.
				WithFields(log.Fields{
					"Bucket": object.Bucket,
					"Object": object.Key,
				}).
				WithError(err).
				Error("Error cleaning file from /temp")
			return err
		}
	}

	// send batch metadata to sqs
	if err := m.SendBatchMetadata(ctx, batchMetadata); err != nil {
		mapperLogger.WithError(err).Error("Error sending metadata to streams")
		return err
	}

	// send event to queue indicating this mapper has completed
	if err := m.SendFinishedEvent(ctx); err != nil {
		mapperLogger.WithError(err).Error("Error sending done event to stream")
		return err
	}

	return nil
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
