// Code generated by ribble DO NOT EDIT.
// |\   \\\\__     o
// | \_/    o \    o
// > _   (( <_  oo
// | / \__+___/
// |/     |/

package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/josenarvaezp/displ/internal/lambdas"
	"github.com/josenarvaezp/displ/pkg/aggregators"
	log "github.com/sirupsen/logrus"
)

var r *lambdas.Reducer

func init() {
	// set logger
	log.SetLevel(log.ErrorLevel)

	var err error
	r, err = lambdas.NewMapMaxReducer(false)
	if err != nil {
		log.WithError(err).Fatal("Error starting reducer")
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.ReducerInput) error {
	// update reducer
	r.UpdateReducerWithRequest(ctx, request)

	// get reduce queue information
	queueName := fmt.Sprintf("%s-%d", r.JobID.String(), request.QueuePartition)
	queueURL := lambdas.GetQueueURL(queueName, r.Region, r.AccountID, r.Local)

	// set reducer logger
	reducerLogger := log.WithFields(log.Fields{
		"Job ID":          r.JobID.String(),
		"Reducer ID":      r.ReducerID.String(),
		"Queue Partition": r.QueuePartition,
	})

	// set wait group
	var wg sync.WaitGroup

	// get checkpoint data
	checkpointData, err := r.GetCheckpointData(ctx, &wg)
	if err != nil {
		reducerLogger.WithError(err).Error("Error reading checkpoint")
		return err
	}

	// batch metadata - number of batches the reducer needs to process
	totalBatchesToProcess, err := r.GetNumberOfBatchesToProcess(ctx)
	if err != nil {
		reducerLogger.WithError(err).Error("Error getting queue metadata")
		return err
	}
	totalProcessedBatches := 0

	// checkpoint info
	processedMessagesWithoutCheckpoint := 0
	checkpointData.LastCheckpoint++

	// holds the intermediate results
	intermediateReducedMap := make(aggregators.MapMax)

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
		WaitTimeSeconds: int32(5),
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
			go r.SaveIntermediateOutput(ctx, intermediateReducedMap, checkpointData.LastCheckpoint, &wg)

			// update output map with reduced intermediate results
			wg.Add(1)
			go r.Output.UpdateOutput(intermediateReducedMap, &wg)

			// delete all messages from queue
			wg.Add(1)
			go r.DeleteIntermediateMessagesFromQueue(ctx, queueURL, processedMessagesDeleteInfo, &wg)

			// update checkpoint info
			checkpointData.LastCheckpoint++
			processedMessagesWithoutCheckpoint = 0
			processedMessagesDeleteInfo = make([]sqsTypes.DeleteMessageBatchRequestEntry, lambdas.MaxMessagesWithoutCheckpoint)
			intermediateReducedMap = make(aggregators.MapMax)
			r.Dedupe.WriteMap = lambdas.InitDedupeMap()
		}

		// call sqs receive messages
		output, err := r.QueuesAPI.ReceiveMessage(ctx, recieveMessageParams)
		if err != nil {
			reducerLogger.WithError(err).Error("Error reading from queue")
			return err
		}

		// process messages
		for _, message := range output.Messages {
			processedMessagesWithoutCheckpoint++

			// add delete info
			processedMessagesDeleteInfo[processedMessagesWithoutCheckpoint] = sqsTypes.DeleteMessageBatchRequestEntry{
				Id:            message.MessageId,
				ReceiptHandle: message.ReceiptHandle,
			}

			// get message attributes
			currentMapID := message.MessageAttributes[lambdas.MapIDAttribute].StringValue
			currentBatchID, err := strconv.Atoi(*message.MessageAttributes[lambdas.BatchIDAttribute].StringValue)
			if err != nil {
				reducerLogger.WithError(err).Error("Error getting message batch ID")
				return err
			}
			currentMessageID, err := strconv.Atoi(*message.MessageAttributes[lambdas.MessageIDAttribute].StringValue)
			if err != nil {
				reducerLogger.WithError(err).Error("Error getting message ID")
				return err
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
			if err := intermediateReducedMap.Reduce(message.Body); err != nil {
				reducerLogger.WithError(err).Error("Error processing message")
				return err
			}
		}

		// check if we are done processing values
		if totalProcessedBatches == *totalBatchesToProcess {
			break
		}
	}

	// wait in case reducers is saving checkpoint in the background
	wg.Wait()

	// update output map with reduced intermediate results
	wg.Add(1)
	go r.Output.UpdateOutput(intermediateReducedMap, &wg)

	// delete all messages from queue
	wg.Add(1)
	go r.DeleteIntermediateMessagesFromQueue(ctx, queueURL, processedMessagesDeleteInfo, &wg)

	wg.Wait()

	// write reducer output
	key := fmt.Sprintf("output/%s", r.ReducerID.String())
	err = r.WriteReducerOutput(ctx, r.Output, key)
	if err != nil {
		reducerLogger.WithError(err).Error("Error writing reducer output")
		return err
	}

	// indicate reducer has finished
	err = r.SendFinishedEvent(ctx)
	if err != nil {
		reducerLogger.WithError(err).Error("Error sending done message")
		return err
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
