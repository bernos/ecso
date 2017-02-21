package commands

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceEventsEnvironmentOption = "environment"
)

func NewServiceEventsCommand(name string) ecso.Command {
	return &serviceEventsCommand{
		name: name,
	}
}

type serviceEventsCommand struct {
	name        string
	environment string
}

func (cmd *serviceEventsCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceDownEnvironmentOption)
	return nil
}

func (cmd *serviceEventsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log       = ctx.Config.Logger()
		env       = ctx.Project.Environments[cmd.environment]
		service   = ctx.Project.Services[cmd.name]
		registry  = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsHelper = helpers.NewECSHelper(registry.ECSAPI(), log.Child())
		ecsoAPI   = api.New(ctx.Config)
		count     = 0
	)

	runningService, err := ecsoAPI.GetECSService(ctx.Project, env, service)

	if err != nil || runningService == nil {
		return err
	}

	cancel := ecsHelper.LogServiceEvents(*runningService.ServiceName, env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
		if err != nil {
			log.Errorf("%s\n", err.Error())
		} else {
			log.Printf("%s %s\n", *e.CreatedAt, *e.Message)
		}
	})

	defer cancel()

	for count < 10 {
		time.Sleep(time.Second * 60)
		count++
	}

	return nil
}

func (cmd *serviceEventsCommand) Validate(ctx *ecso.CommandContext) error {
	err := util.AnyError(
		ui.ValidateRequired("Name")(cmd.name),
		ui.ValidateRequired("Environment")(cmd.name))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[cmd.name]; !ok {
		return fmt.Errorf("Service '%s' not found", cmd.name)
	}

	if _, ok := ctx.Project.Environments[cmd.environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", cmd.environment)
	}

	return nil
}

func (cmd *serviceEventsCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
