package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/lambdas"
)

var r *lambdas.Reducer

func init() {
	var err error
	r, err = lambdas.NewReducer(true)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.ReducerInput) (string, error) {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return "", errors.New("Error getting lambda context")
	}
	r.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	r.ReducerID = uuid.New()
	r.JobID = request.JobID
	r.NumMappers = request.NumMappers
	r.QueuePartition = request.QueuePartition

	queueName := fmt.Sprintf("%s-%d", r.JobID.String(), request.QueuePartition)
	queueURL := lambdas.GetQueueURL(queueName, r.Region, r.AccountID, r.Local)

	// batch metadata - number of batches the reducer needs to process
	totalBatchesToProcess, err := r.GetNumberOfBatchesToProcess(ctx)
	totalProcessedBatches := 0

	// map to hold data of all processed messages
	dedupeMap := lambdas.InitDedupeMap()

	// init output map - holds reduced values
	outputMap := make(map[string]int)

	// use same parameters for all get messages requests
	recieveMessageParams := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: lambdas.MaxItemsPerBatch,
		MessageAttributeNames: []string{
			lambdas.MapIDAttribute,
			lambdas.BatchIDAttribute,
			lambdas.MessageIDAttribute,
		},
	}

	// recieve messages until we are done processing all queue
	for true {
		output, err := r.QueuesAPI.ReceiveMessage(ctx, recieveMessageParams)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		for _, message := range output.Messages {
			// unmarshall message body
			var res lambdas.MapInt
			body := []byte(*message.Body)
			err = json.Unmarshal(body, &res)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			// get message attributes
			currentMapID := message.MessageAttributes[lambdas.MapIDAttribute].StringValue
			currentBatchID, err := strconv.Atoi(*message.MessageAttributes[lambdas.BatchIDAttribute].StringValue)
			if err != nil {
				fmt.Println(err)
				return "", err
			}
			currentMessageID, err := strconv.Atoi(*message.MessageAttributes[lambdas.MessageIDAttribute].StringValue)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			// check if message has already been processed
			dedupeMessages, ok := dedupeMap.GetProcessedMessages(*currentMapID, currentBatchID)
			if ok {
				if dedupeMessages.IsBatchComplete() {
					// ignore as it is a duplicated message
					continue
				}

				if dedupeMessages.IsMessageProcessed(currentMessageID) {
					// ignore as it is a duplicated message
					continue
				}

				// message has not been processed
				// add processed message to dedupe map
				dedupeMessages.UpdateMessageProcessed(currentMessageID)

				// check if we are done processing batch from map
				if dedupeMessages.IsBatchComplete() {
					totalProcessedBatches++
					// delete processed map
					dedupeMessages.DeletedProcessedMessages()
				}
			} else {
				// no messages for batch have been processed - init dedupe data for batch
				dedupeMap.InitDedupeBatch(*currentMapID, currentBatchID, currentMessageID)
			}

			// process message
			currentKey := res.Key
			currentValue := res.Value

			// only process value if it is not empty
			// emty values are sent to keep the same number of events per batch
			if res.EmptyVal != true {
				outputMap[currentKey] = outputMap[currentKey] + currentValue
			}
		}

		// check if we are done processing values
		if totalProcessedBatches == *totalBatchesToProcess {
			break
		}
	}

	err = r.WriteReducerOutput(ctx, outputMap)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return "", nil
}

func main() {
	ctx := context.Background()
	request := lambdas.ReducerInput{
		JobID:          uuid.MustParse("1469d3b8-d133-4036-8944-01ff6518ec25"),
		QueuePartition: 4,
	}
	HandleRequest(ctx, request)
	// lambda.Start(HandleRequest)
}
