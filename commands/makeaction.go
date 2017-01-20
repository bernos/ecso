package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(fn func(*cli.Context) ecso.Command, dispatcher ecso.Dispatcher, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := dispatcher.Dispatch(fn(c), options...); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	}
}
