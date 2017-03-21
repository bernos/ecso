package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

// NewApp creates a new `cli.App` interface for the ecso command line utility
func NewApp(cfg *ecso.Config, dispatcher ecso.Dispatcher) *cli.App {
	app := cli.NewApp()
	app.Name = "ecso"
	app.Usage = "Manage Amazon ECS projects"
	app.Version = cfg.Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Brendan McMahon",
			Email: "bernos@gmail.com",
		},
	}

	cli.ErrWriter = cfg.Logger().ErrWriter()

	cliDispatcher := Dispatcher(dispatcher)

	app.Commands = []cli.Command{
		NewInitCliCommand(cliDispatcher),
		NewEnvironmentCliCommand(cliDispatcher),
		NewServiceCliCommand(cliDispatcher),
		NewEnvCliCommand(cliDispatcher),
	}

	return app
}

// factory is a function that creates an `ecso.Command` from a `cli.Context`
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

// Dispatcher wraps a dispatcher and returns a `cli.ExitError` if the underlying
// dipatcher fails
func Dispatcher(dispatcher ecso.Dispatcher) ecso.Dispatcher {
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
