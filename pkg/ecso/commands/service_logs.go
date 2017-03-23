package commands

import (
	"time"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

func NewServiceLogsCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &serviceLogsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceLogsCommand struct {
	*ServiceCommand
}

func (cmd *serviceLogsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	events, err := cmd.serviceAPI.ServiceLogs(ctx.Project, env, service)

	if err != nil {
		return err
	}

	for _, event := range events {
		cmd.log.Printf("%-42s %s\n", time.Unix(*event.Timestamp/1000, *event.Timestamp%1000), *event.Message)
	}

	return nil
}
