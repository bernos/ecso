package environmentup

import (
	"os"

	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	DryRun string
}{
	DryRun: "dry-run",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:        "up",
		Usage:       "Create/update an ecso environment",
		Description: "TODO",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.DryRun,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
		},
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.Args().First()

	if env == "" {
		env = os.Getenv("ECSO_ENVIRONMENT")
	}

	if env == "" {
		return nil, cmd.NewArgumentRequiredError("environment")
	}

	return commands.NewEnvironmentUpCommand(env, func(opt *commands.EnvironmentUpOptions) {
		opt.DryRun = c.Bool(keys.DryRun)
	}), nil
}
