package environmentup

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name   string
	DryRun string
}{
	Name:   "name",
	DryRun: "dry-run",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "up",
		Usage: "Bring up the named environment",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Name,
				Usage:  "The name of the environment to bring up. If the environment doesn't exist it will be created, otherwise it will be updated.",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.BoolFlag{
				Name:  keys.DryRun,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
		},
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.String(keys.Name)

	if env == "" {
		return nil, cmd.NewOptionRequiredError(keys.Name)
	}

	return commands.NewEnvironmentUpCommand(env, func(opt *commands.EnvironmentUpOptions) {
		opt.DryRun = c.Bool(keys.DryRun)
	}), nil
}
