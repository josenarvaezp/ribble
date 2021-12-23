package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/josenarvaezp/displ/internal/lambdas"
)

var c *lambdas.Coordinator

func init() {
	var err error
	c, err = lambdas.NewCoordinator(true)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func HandleRequest(ctx context.Context, request lambdas.CoordinatorInput) (string, error) {
	// get data from context
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return "", errors.New("Error getting lambda context")
	}
	c.AccountID = strings.Split(lc.InvokedFunctionArn, ":")[4]
	c.JobID = request.JobID
	c.NumMappers = int64(request.NumMappers)
	c.NumQueues = int64(request.NumQueues)

	// waits until mappers are done
	if err := c.AreMappersDone(ctx); err != nil {
		fmt.Println(err)
		return "", err
	}

	// invoke reducers
	if err := c.InvokeReducers(ctx); err != nil {
		fmt.Println(err)
		return "", nil
	}

	return "", nil
}

func main() {
	lambda.Start(HandleRequest)
}
