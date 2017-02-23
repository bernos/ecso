package commands

import (
	"time"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceLogsCommand(name string) ecso.Command {
	return &serviceLogsCommand{
		ServiceCommand: &ServiceCommand{
			name: name,
		},
	}
}

type serviceLogsCommand struct {
	*ServiceCommand
}

func (cmd *serviceLogsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		service = ctx.Project.Services[cmd.name]
		env     = ctx.Project.Environments[cmd.environment]
		log     = ctx.Config.Logger()
		ecsoAPI = api.New(ctx.Config)
	)

	events, err := ecsoAPI.ServiceLogs(ctx.Project, env, service)

	if err != nil {
		return err
	}

	for _, event := range events {
		log.Printf("%-42s %s\n", time.Unix(*event.Timestamp/1000, *event.Timestamp%1000), *event.Message)
	}

	return nil
}
