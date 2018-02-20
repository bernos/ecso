package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		DryRun cli.BoolFlag
		Force  cli.BoolFlag
	}{

		DryRun: cli.BoolFlag{
			Name:  "dry-run",
			Usage: "If set, list pending changes, but do not execute the updates.",
		},
		Force: cli.BoolFlag{
			Name:  "force",
			Usage: "Override warnings about first time environment deployments if cloud formation stack already exists",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentUpCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithDryRun(ctx.Bool(flags.DryRun.Name)).
				WithForce(ctx.Bool(flags.Force.Name))
		})
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploys the infrastructure for an ecso environment",
		Description: "All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.",
		ArgsUsage:   "ENVIRONMENT",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.DryRun,
			flags.Force,
		},
	}
}
