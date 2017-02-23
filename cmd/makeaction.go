package cmd

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

type factory func(*cli.Context) (ecso.Command, error)

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(dispatcher ecso.Dispatcher, fn factory, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(c *cli.Context) error {
		command, err := fn(c)

		if err != nil {
			cli.ShowSubcommandHelp(c)
			return cli.NewExitError(err.Error(), 1)
		}

		if err := command.UnmarshalCliContext(c); err != nil {
			cli.ShowSubcommandHelp(c)
			return cli.NewExitError(err.Error(), 1)
		}

		if err := dispatcher.Dispatch(command, options...); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	}
}
