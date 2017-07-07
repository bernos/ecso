package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/config"
	"gopkg.in/urfave/cli.v1"
)

// NewApp creates a new `cli.App` interface for the ecso command line utility
func NewApp(cfg *config.Config, dispatcher ecso.Dispatcher) *cli.App {
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
type factory func(*cli.Context, *config.Config) (ecso.Command, error)

func CommandFactory(ctx *cli.Context, fn factory) ecso.CommandFactory {
	return func(cfg *config.Config) (ecso.Command, error) {
		return fn(ctx, cfg)
	}
}

// Dispatcher wraps a dispatcher and returns a `cli.ExitError` if the underlying
// dipatcher fails
func Dispatcher(dispatcher ecso.Dispatcher) ecso.Dispatcher {
	return ecso.DispatcherFunc(func(factory ecso.CommandFactory, cOptions ecso.CommandOptions, options ...func(*ecso.DispatchOptions)) error {
		if err := dispatcher.Dispatch(factory, cOptions, options...); err != nil {
			if ecso.IsArgumentRequiredError(err) || ecso.IsOptionRequiredError(err) {
				cli.ShowSubcommandHelp(cOptions.(*cli.Context))
			}

			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	})
}

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(dispatcher ecso.Dispatcher, fn factory, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		return dispatcher.Dispatch(CommandFactory(ctx, fn), ctx, options...)
	}
}
