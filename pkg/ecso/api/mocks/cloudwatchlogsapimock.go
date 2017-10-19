package mocks

import "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

type CloudWatchLogsAPIMock struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
}
