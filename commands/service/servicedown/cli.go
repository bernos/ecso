package servicedown

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
		Name:  "down",
		Usage: "terminates a service",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "The name of the service to terminate",
			},
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	service := c.String(keys.Name)
	env := c.String(keys.Environment)

	if service == "" {
		return nil, commands.NewOptionRequiredError(keys.Name)
	}

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(c.String(keys.Name), c.String(keys.Environment), func(opt *Options) {
		// TODO: populate options from c
	}), nil
}
