package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

type factory func(*cli.Context) (ecso.Command, error)

// BuildCommand builds an `ecso.Command` using `cli.Context` and an ecso
// Command factory. If there is an error running the factory, or if the cli Context
// cannot be unmarshalled into the Command, a `cli.ExitError` will be returned
func BuildCommand(ctx *cli.Context, fn factory) (ecso.Command, error) {
	command, err := fn(ctx)

	if err != nil {
		return nil, cli.NewExitError(err.Error(), 1)
	}

	if err := command.UnmarshalCliContext(ctx); err != nil {
		return nil, cli.NewExitError(err.Error(), 1)
	}

	return command, nil
}

// CliDispatcher wraps a dispatcher and returns a `cli.ExitError` if the underlying dipatcher
// fails
func CliDispatcher(dispatcher ecso.Dispatcher) ecso.Dispatcher {
	return ecso.DispatcherFunc(func(command ecso.Command, options ...func(*ecso.DispatchOptions)) error {
		if err := dispatcher.Dispatch(command, options...); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	})
}

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(dispatcher ecso.Dispatcher, fn factory, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(c *cli.Context) error {
		command, err := BuildCommand(c, fn)

		if err != nil {
			cli.ShowSubcommandHelp(c)
			return err
		}

		return dispatcher.Dispatch(command, options...)
	}
}
