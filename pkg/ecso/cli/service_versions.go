package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceVersionsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "The name of the environment",
			EnvVar: "ECSO_ENVIRONMENT",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceVersionsCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
		})
	}

	return cli.Command{
		Name:      "versions",
		Usage:     "Show available versions for a service",
		ArgsUsage: "SERVICE",
		Action:    MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
		},
	}
}
