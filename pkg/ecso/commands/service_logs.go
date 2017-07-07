package commands

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

func NewServiceLogsCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceLogsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceLogsCommand struct {
	*ServiceCommand
}

func (cmd *serviceLogsCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	events, err := cmd.serviceAPI.ServiceLogs(ctx.Project, cmd.Environment(ctx), cmd.Service(ctx))

	if err != nil {
		return err
	}

	for _, event := range events {
		l.Printf("%-42s %s\n", cmd.EventTime(event), *event.Message)
	}

	return nil
}

func (cmd *serviceLogsCommand) EventTime(e *cloudwatchlogs.FilteredLogEvent) time.Time {
	return time.Unix(*e.Timestamp/1000, *e.Timestamp%1000)
}
