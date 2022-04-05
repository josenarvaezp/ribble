package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/sirupsen/logrus"
)

// ReadRibbleLogs reads the coordinator logs and prints
// them the as logs
func (d *Driver) ReadRibbleLogs(ctx context.Context, sleepTime int32) error {
	logGroupName := fmt.Sprintf("%s-log-group", d.JobID.String())
	logStreamName := fmt.Sprintf("%s-log-stream", d.JobID.String())

	var nextToken *string

	for true {
		out, err := d.LogsAPI.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  &logGroupName,
			LogStreamName: &logStreamName,
			StartFromHead: aws.Bool(true),
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}

		nextToken = out.NextForwardToken

		for _, event := range out.Events {
			timestamp := time.Unix(*event.Timestamp, 0)
			logrus.WithFields(logrus.Fields{
				"Timestamp": timestamp,
			}).Info(*event.Message)
		}

		// sleep before fetching more logs
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return nil
}
