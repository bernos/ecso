package events

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name        string
	Environment string
}{
	Name:        "name",
	Environment: "environment",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "events",
		Usage: "List events for a service",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "The service to list events for",
			},
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	name := c.String(keys.Name)
	env := c.String(keys.Environment)

	if name == "" {
		return nil, commands.NewOptionRequiredError(keys.Name)
	}

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(name, env, func(opt *Options) {
		// TODO: populate options from c
	}), nil
}

type Options struct {
	Name        string
	Environment string
}

func New(name, env string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:        name,
		Environment: env,
	}

	for _, option := range options {
		option(o)
	}

	return &command{
		options: o,
	}
}

type command struct {
	options *Options
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	if err := validateOptions(cmd.options, ctx); err != nil {
		return err
	}

	var (
		log        = ctx.Config.Logger
		env        = ctx.Project.Environments[cmd.options.Environment]
		service    = ctx.Project.Services[cmd.options.Name]
		registry   = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsService = registry.ECSService(log.PrefixPrintf("  "))
		count      = 0
	)

	cancel := ecsService.LogServiceEvents(service.GetECSServiceName(), env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
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

func validateOptions(opt *Options, ctx *ecso.CommandContext) error {
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
