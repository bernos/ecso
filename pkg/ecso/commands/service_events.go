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
)

type ServiceEventsOptions struct {
	Name        string
	Environment string
}

func NewServiceEventsCommand(name, env string, options ...func(*ServiceEventsOptions)) ecso.Command {
	o := &ServiceEventsOptions{
		Name:        name,
		Environment: env,
	}

	for _, option := range options {
		option(o)
	}

	return &serviceEventsCommand{
		options: o,
	}
}

type serviceEventsCommand struct {
	options *ServiceEventsOptions
}

func (cmd *serviceEventsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log       = ctx.Config.Logger()
		env       = ctx.Project.Environments[cmd.options.Environment]
		service   = ctx.Project.Services[cmd.options.Name]
		registry  = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsHelper = helpers.NewECSHelper(registry.ECSAPI(), log.PrefixPrintf("  "))
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
	opt := cmd.options

	err := util.AnyError(
		ui.ValidateRequired("Name")(opt.Name),
		ui.ValidateRequired("Environment")(opt.Environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[opt.Name]; !ok {
		return fmt.Errorf("Service '%s' not found", opt.Name)
	}

	if _, ok := ctx.Project.Environments[opt.Environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
	}

	return nil
}

func (cmd *serviceEventsCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
