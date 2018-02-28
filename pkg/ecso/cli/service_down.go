package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceDownCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
		Force       cli.BoolFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "The environment to terminate the service from",
			EnvVar: "ECSO_ENVIRONMENT",
		},
		Force: cli.BoolFlag{
			Name:  "force",
			Usage: "Required. Confirms the service will be terminated",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceDownCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region)).
				WithForce(ctx.Bool(flags.Force.Name))
		})
	}

	return cli.Command{
		Name:        "down",
		Usage:       "terminates a service",
		Description: "The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
			flags.Force,
		},
	}
}
