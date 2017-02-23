package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceDescribeCommand(name string) ecso.Command {
	return &serviceDecribeCommand{
		ServiceCommand: &ServiceCommand{
			name: name,
		},
	}
}

type serviceDecribeCommand struct {
	*ServiceCommand
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.environment]
		service = ctx.Project.Services[cmd.name]
		log     = ctx.Config.Logger()
		ecsoAPI = api.NewServiceAPI(ctx.Config)
	)

	description, err := ecsoAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(log, description)

	return nil
}
