package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	log "github.com/sirupsen/logrus"

	"github.com/josenarvaezp/displ/internal/lambdas"
	"github.com/josenarvaezp/displ/pkg/aggregators"
)

var r *lambdas.Reducer

func init() {
	// set logger
	log.SetLevel(log.ErrorLevel)

	var err error
	r, err = lambdas.NewRandomReducer(false)
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
	totalMessagesToProcess, err := r.GetNumberOfBatchesToProcess(ctx)
	if err != nil {
		reducerLogger.WithError(err).Error("Error getting queue metadata")
		return err
	}
	totalProcessedMessages := 0

	// checkpoint info
	processedMessagesWithoutCheckpoint := 0
	checkpointData.LastCheckpoint++

	// holds the intermediate results
	intermediateReducedMap := make(aggregators.MapAggregator)

	// processedMessagesDeleteInfo holds the data to delete messages from queue
	processedMessagesDeleteInfo := make([]sqsTypes.DeleteMessageBatchRequestEntry, lambdas.MaxMessagesWithoutCheckpoint)

	// use same parameters for all get messages requests
	recieveMessageParams := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: lambdas.MaxItemsPerBatch,
		MessageAttributeNames: []string{
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
			r.DedupeSimple.Merge()

			// save intermediate dedupe
			wg.Add(1)
			go r.SaveIntermediateDedupe(ctx, checkpointData.LastCheckpoint, r.DedupeSimple.WriteMap, &wg)

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
			intermediateReducedMap = make(aggregators.MapAggregator)
			r.DedupeSimple.WriteMap = lambdas.InitDedupeSimpleMap()
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
			currentMessageID := *message.MessageAttributes[lambdas.MessageIDAttribute].StringValue

			// check if message has already been processed
			if !r.DedupeSimple.IsMessageProcessed(currentMessageID) {

				// process message
				// unmarshall message body
				var reduceMessage *aggregators.ReduceMessage
				body := []byte(*message.Body)
				err = json.Unmarshal(body, &reduceMessage)
				if err != nil {
					return err
				}

				// process message
				if err := intermediateReducedMap.Reduce(reduceMessage); err != nil {
					reducerLogger.WithError(err).Error("Error processing message")
					return err
				}

				// update dedupe and messages processed count
				r.DedupeSimple.UpdateMessageProcessed(currentMessageID)
				totalProcessedMessages++
			}
		}

		// check if we are done processing values
		if totalProcessedMessages == *totalMessagesToProcess {
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
	messagesSent, err := r.EmitValuesToFinalReducer(ctx)
	if err != nil {
		reducerLogger.WithError(err).Error("Error sending reducer output to final reduce queue")
		return err
	}

	// send message metadata to sqs
	if err := r.SendMetadata(ctx, messagesSent); err != nil {
		reducerLogger.WithError(err).Error("Error sending metadata to final stream")
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
