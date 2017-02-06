package ps

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"

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
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	name := c.String(keys.Name)
	env := c.String(keys.Environment)

	if name == "" {
		return nil, cmd.NewOptionRequiredError(keys.Name)
	}

	if env == "" {
		return nil, cmd.NewOptionRequiredError(keys.Environment)
	}

	return commands.NewServicePsCommand(name, env, func(opt *commands.ServicePsOptions) {
		// TODO: populate options from c
	}), nil
}
