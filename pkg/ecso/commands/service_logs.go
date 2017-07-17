package commands

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceLogsCommand(name string, environmentName string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceLogsCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type serviceLogsCommand struct {
	*ServiceCommand
}

func (cmd *serviceLogsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	events, err := cmd.serviceAPI.ServiceLogs(ctx.Project, cmd.Environment(ctx), cmd.Service(ctx))

	if err != nil {
		return err
	}

	for _, event := range events {
		fmt.Fprintf(w, "%-42s %s\n", cmd.EventTime(event), *event.Message)
	}

	return nil
}

func (cmd *serviceLogsCommand) EventTime(e *cloudwatchlogs.FilteredLogEvent) time.Time {
	return time.Unix(*e.Timestamp/1000, *e.Timestamp%1000)
}
