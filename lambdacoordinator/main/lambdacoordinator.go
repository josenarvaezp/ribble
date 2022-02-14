package main

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	log "github.com/sirupsen/logrus"

	"github.com/josenarvaezp/displ/internal/lambdas"
	"github.com/josenarvaezp/displ/pkg/aggregators"
)

var c *lambdas.Coordinator

func init() {
	// set logger
	log.SetLevel(log.ErrorLevel)

	var err error
	c, err = lambdas.NewCoordinator(true)
	if err != nil {
		log.WithError(err).Fatal("Error starting coordinator")
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.CoordinatorInput) error {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return errors.New("Error getting lambda context")
	}
	c.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	c.JobID = request.JobID
	c.NumMappers = int64(request.NumMappers)
	c.NumQueues = int64(request.NumQueues)

	coordinatorLogger := log.WithFields(log.Fields{
		"Job ID": c.JobID.String(),
	})

	// waits until mappers are done
	if err := c.AreMappersDone(ctx); err != nil {
		coordinatorLogger.WithError(err).Error("Error reading mappers done queue")
		return err
	}

	// invoke reducers
	if err := c.InvokeReducers(ctx, aggregators.AggregatorMapSum); err != nil {
		coordinatorLogger.WithError(err).Error("Error invoking reducers")
		return nil
	}

	// wait until reducers are done
	if err := c.AreReducersDone(ctx); err != nil {
		coordinatorLogger.WithError(err).Error("Error reading reducers done queue")
		return err
	}

	// indicate reducers are done
	if err := c.WriteDoneObject(ctx); err != nil {
		coordinatorLogger.WithError(err).Error("Error writing done signal")
		return err
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
