package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceRollbackCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
		Version     cli.StringFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "The name of the environment to deploy to",
			EnvVar: "ECSO_ENVIRONMENT",
		},
		Version: cli.StringFlag{
			Name:  "version",
			Usage: "The version to rollback to",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceRollbackCommand(
				service.Name,
				env.Name,
				ctx.String(flags.Version.Name),
				cfg.ServiceAPI(env.Region))
		})
	}

	return cli.Command{
		Name:        "rollback",
		Usage:       "Rollback a service to an earlier version",
		Description: "Replace the currently running service with a previously deployed service version",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
			flags.Version,
		},
	}
}
