package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceDescribeCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "The environment to query",
			EnvVar: "ECSO_ENVIRONMENT",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceDescribeCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
		})
	}

	return cli.Command{
		Name:        "describe",
		Usage:       "Lists details of a deployed service",
		Description: "Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
		},
	}
}
