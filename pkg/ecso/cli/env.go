package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Unset cli.BoolFlag
	}{
		Unset: cli.BoolFlag{
			Name:  "unset",
			Usage: "If set, output shell commands to unset all ecso environment variables",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewEnvCommand(ctx.Args().First()).
			WithUnset(ctx.Bool(flags.Unset.Name)), nil
	}

	return cli.Command{
		Name:      "env",
		Usage:     "Display the commands to set up the default environment for the ecso cli tool",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Unset,
		},
	}
}
