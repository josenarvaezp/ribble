package lambdas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/josenarvaezp/displ/internal/config"
	"github.com/josenarvaezp/displ/internal/faas"
	"github.com/josenarvaezp/displ/internal/queues"
)

const (
	CoordinatorName = "displ-coordinator" // TODO: name of the function or ARN
)

// CoordinatorInput is the input the coordinator lambda receives
type CoordinatorInput struct {
	JobID      uuid.UUID `json:"jobID"`
	NumMappers int       `json:"numMappers"`
	NumQueues  int       `json:"numQueues"`
}

// CoordinatorAPI is an interface deining the functions available to the coordinator
type CoordinatorAPI interface {
	AreMappersDone(ctx context.Context) (bool, error)
	InvokeReducers(ctx context.Context) error
}

// Coordinator is an interface that implements CoordinatorAPI
type Coordinator struct {
	JobID uuid.UUID
	// clients
	QueuesAPI queues.QueuesAPI
	FaasAPI   faas.FaasAPI
	// metadata
	Region     string
	AccountID  string
	NumMappers int64
	NumQueues  int64
	local      bool
}

// NewCoordinator initializes a new coordinator with its required clients
func NewCoordinator(
	local bool,
) (*Coordinator, error) {
	var cfg aws.Config
	var err error

	// get region from env var
	region := os.Getenv("AWS_REGION")

	// init coordinator
	coordinator := &Coordinator{
		Region: region,
		local:  local,
	}

	// create config
	if local {
		cfg, err = config.InitLocalCfg()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.InitCfg(region)
		if err != nil {
			return nil, err
		}
	}

	// create sqs client
	coordinator.QueuesAPI = sqs.NewFromConfig(cfg)

	// create lambda client
	coordinator.FaasAPI = lambda.NewFromConfig(cfg)

	return coordinator, err
}

// AreMappersDone reads events from the mapper-done queue to check
// if all mappers are done
func (c *Coordinator) AreMappersDone(ctx context.Context) error {
	queueName := fmt.Sprintf("%s-mappers-done", c.JobID.String())
	queueURL := GetQueueURL(queueName, c.Region, c.AccountID, c.local)
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: MaxItemsPerBatch,
	}

	// keeps a map of done mappers, this is used as the dedupe mechanism
	doneMappers := make(map[string]bool)
	doneMappersCount := 0

	// loop until all mappers are done
	for doneMappersCount < int(c.NumMappers) {
		// mappers are not done yet
		output, err := c.QueuesAPI.ReceiveMessage(ctx, params)
		if err != nil {
			return err
		}

		for _, message := range output.Messages {
			// add mapper to done map
			if _, ok := doneMappers[*message.Body]; !ok {
				doneMappers[*message.Body] = true
				doneMappersCount++
			}
		}

		// sleep for 10 seconds before trying to get more results
		time.Sleep(10 * time.Second)
	}

	return nil
}

// InvokeReducers is used to invoke the reducers once all mappers are done
// there is one reducer per queue invoked
func (c *Coordinator) InvokeReducers(ctx context.Context) error {
	// invoke a reducer per each queue
	for i := 0; i < int(c.NumQueues); i++ {
		// encode reducer input to json
		reducerInput := ReducerInput{
			JobID:          c.JobID,
			ReducerID:      uuid.New(),
			QueuePartition: i,
			NumMappers:     int(c.NumMappers),
		}
		requestPayload, err := json.Marshal(reducerInput)
		if err != nil {
			return err
		}

		result, _ := c.FaasAPI.Invoke(
			ctx,
			&lambda.InvokeInput{
				FunctionName:   aws.String(CoordinatorName),
				Payload:        requestPayload,
				InvocationType: types.InvocationTypeEvent,
			},
		)

		// error is ignored from asynch invokation and result only holds the status code
		// check status code
		if result.StatusCode != 202 { //SUCCESS_CODE
			// TODO: stop execution and inform the user about the errors
			return errors.New("Error starting mappers")
		}
	}

	return nil
}
