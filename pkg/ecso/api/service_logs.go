package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/bernos/ecso/pkg/ecso"
)

func (api *api) ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cwLogsAPI := reg.CloudWatchLogsAPI()

	resp, err := cwLogsAPI.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(s.GetCloudWatchLogGroupName(env)),
	})

	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}
