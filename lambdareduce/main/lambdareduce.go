package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
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
	r.ReducerID = request.ReducerID
	r.JobID = request.JobID
	r.NumMappers = request.NumMappers
	r.QueuePartition = request.QueuePartition

	var wg sync.WaitGroup

	// check if there is a checkpoint saved for this reducer
	checkpointData, err := r.GetCheckpointData(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// get output map and dedupe info from checkpoints
	if len(checkpointData.IntermediateOutputData) != 0 {
		r.GetOutputMap(ctx, checkpointData.IntermediateOutputData, &wg)
		r.GetDedupe(ctx, checkpointData.DedupeData, &wg)

		wg.Wait()
	}

	queueName := fmt.Sprintf("%s-%d", r.JobID.String(), request.QueuePartition)
	queueURL := lambdas.GetQueueURL(queueName, r.Region, r.AccountID, r.Local)

	// batch metadata - number of batches the reducer needs to process
	totalBatchesToProcess, err := r.GetNumberOfBatchesToProcess(ctx)
	totalProcessedBatches := 0

	// checkpoint info
	processedMessagesWithoutCheckpoint := 0
	checkpointData.LastCheckpoint++

	// holds the intermediate results
	intermediateMap := make(map[string]int)

	// processedMessagesDeleteInfo holds the data to delete messages from queue
	processedMessagesDeleteInfo := make([]sqsTypes.DeleteMessageBatchRequestEntry, lambdas.MaxMessagesWithoutCheckpoint)

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
		if processedMessagesWithoutCheckpoint == lambdas.MaxMessagesBeforeCheckpointComplete && checkpointData.LastCheckpoint != 1 {
			// check that the last checkpoint has completed before processing any more messages
			// we give a buffer of 15,000 new messages for saving the checkpoint which happens
			// in the background. If this point is reached it means we have processed 115,000 messages
			// without deleting from the queue which is close to the aws limit for queues
			wg.Wait()
		}

		if processedMessagesWithoutCheckpoint == lambdas.MaxMessagesWithoutCheckpoint {
			// We need to delete the messages read from the sqs queue and we create a checkpoint
			// in S3 as the fault tolerant mechanism. Saving the checkpoint can be done concurrently
			// in the background while we keep processing messages

			// merge the dedupe map so that the read dedupe map is up to date
			r.Dedupe.Merge()

			// save intermediate dedupe
			wg.Add(1)
			go r.SaveIntermediateDedupe(ctx, checkpointData.LastCheckpoint, &wg)

			// save intermediate map
			wg.Add(1)
			go r.SaveIntermediateOutput(ctx, intermediateMap, checkpointData.LastCheckpoint, &wg)

			// update output map with reduced intermediate results
			wg.Add(1)
			go r.UpdateOutputMap(intermediateMap, &wg)

			// delete all messages from queue
			wg.Add(1)
			go r.DeleteIntermediateMessagesFromQueue(ctx, queueURL, processedMessagesDeleteInfo, &wg)

			// update checkpoint info
			checkpointData.LastCheckpoint++
			processedMessagesWithoutCheckpoint = 0
			processedMessagesDeleteInfo = make([]sqsTypes.DeleteMessageBatchRequestEntry, lambdas.MaxMessagesWithoutCheckpoint)
			intermediateMap = make(map[string]int)
			r.Dedupe.WriteMap = lambdas.InitDedupeMap()
		}

		output, err := r.QueuesAPI.ReceiveMessage(ctx, recieveMessageParams)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		for _, message := range output.Messages {
			processedMessagesWithoutCheckpoint++

			// add delete info
			processedMessagesDeleteInfo[processedMessagesWithoutCheckpoint] = sqsTypes.DeleteMessageBatchRequestEntry{
				Id:            message.MessageId,
				ReceiptHandle: message.ReceiptHandle,
			}

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
			if exists := r.Dedupe.BatchExists(*currentMapID, currentBatchID); exists {
				if r.Dedupe.IsBatchComplete(*currentMapID, currentBatchID) {
					// ignore as it is a duplicated message
					continue
				}

				if r.Dedupe.IsMessageProcessed(*currentMapID, currentBatchID, currentMessageID) {
					// ignore as it is a duplicated message
					continue
				}

				// message has not been processed
				// add processed message to dedupe map
				r.Dedupe.UpdateMessageProcessed(*currentMapID, currentBatchID, currentMessageID)

				// check if we are done processing batch from map
				if r.Dedupe.IsBatchComplete(*currentMapID, currentBatchID) {
					totalProcessedBatches++
					// delete processed map from dedupe
					r.Dedupe.DeletedProcessedMessages(*currentMapID, currentBatchID)
				}
			} else {
				// no messages for batch have been processed - init dedupe data for batch
				r.Dedupe.InitDedupeBatch(*currentMapID, currentBatchID, currentMessageID)
			}

			// process message
			currentKey := res.Key
			currentValue := res.Value

			// only process value if it is not empty
			// empty values are sent to keep the same number of events per batch
			if res.EmptyVal != true {
				intermediateMap[currentKey] = intermediateMap[currentKey] + currentValue
			}
		}

		// check if we are done processing values
		if totalProcessedBatches == *totalBatchesToProcess {
			break
		}
	}

	// wait in case reducers is saving checkpoint in the background
	wg.Wait()

	// save intermediate dedupe
	wg.Add(1)
	go r.SaveIntermediateDedupe(ctx, checkpointData.LastCheckpoint, &wg)

	// save intermediate map
	wg.Add(1)
	go r.SaveIntermediateOutput(ctx, intermediateMap, checkpointData.LastCheckpoint, &wg)

	// update output map with reduced intermediate results
	wg.Add(1)
	go r.UpdateOutputMap(intermediateMap, &wg)

	// delete all messages from queue
	wg.Add(1)
	go r.DeleteIntermediateMessagesFromQueue(ctx, queueURL, processedMessagesDeleteInfo, &wg)

	wg.Wait()

	// write reducer output
	key := fmt.Sprintf("output/%s", r.ReducerID.String())
	err = r.WriteReducerOutput(ctx, r.OutputMap, key)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// indicate reducer has finished
	err = r.SendFinishedEvent(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
