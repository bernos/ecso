package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentDescribeCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentDescribeCommand(env.Name, cfg.EnvironmentAPI(env.Region))
		})
	}

	return cli.Command{
		Name:      "describe",
		Usage:     "Describes an ecso environment",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fn),
	}
}
