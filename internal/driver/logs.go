package driver

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func (d *Driver) GetLogs(ctxt context.Context) {
	aggregatorGroupName := "/aws/lambda/map_aggregator"
	d.LogsAPI.GetLogEvents(ctxt, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName: &aggregatorGroupName,
	})
}
