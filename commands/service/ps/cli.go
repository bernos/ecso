package ps

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Environment string
}{
	Environment: "environment",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:        "ps",
		Usage:       "Show running tasks for a service",
		Description: "TODO",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
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
	name := c.Args().First()
	env := c.String(keys.Environment)

	if name == "" {
		return nil, commands.NewArgumentRequiredError("service")
	}

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(name, env, func(opt *Options) {
		// TODO: populate options from c
	}), nil
}
