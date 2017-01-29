package ps

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name        string
	Environment string
}{
	Name:        "name",
	Environment: "environment",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "ps",
		Usage: "Show running tasks for a service",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "The service to list tasks for",
			},
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
	name := c.String(keys.Name)
	env := c.String(keys.Environment)

	if name == "" {
		return nil, commands.NewOptionRequiredError(keys.Name)
	}

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(name, env, func(opt *Options) {
		// TODO: populate options from c
	}), nil
}
