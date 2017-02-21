package commands

import (
	"fmt"
	"time"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceLogsEnvironmentOption = "environment"
)

func NewServiceLogsCommand(name string) ecso.Command {
	return &serviceLogsCommand{
		name: name,
	}
}

type serviceLogsCommand struct {
	name        string
	environment string
}

func (cmd *serviceLogsCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceLogsEnvironmentOption)
	return nil
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

func (cmd *serviceLogsCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.name == "" {
		return fmt.Errorf("Name is required")
	}

	if cmd.environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(cmd.name) {
		return fmt.Errorf("No service named '%s' was found", cmd.name)
	}

	if !ctx.Project.HasEnvironment(cmd.environment) {
		return fmt.Errorf("No environment named '%s' was found", cmd.environment)
	}

	return nil
}

func (cmd *serviceLogsCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
