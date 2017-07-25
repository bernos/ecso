package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

// NewApp creates a new `cli.App` interface for the ecso command line utility
func NewApp(cfg *config.Config, project *ecso.Project, dispatcher dispatcher.Dispatcher) *cli.App {
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

	cli.ErrWriter = cfg.ErrWriter()

	app.Commands = []cli.Command{
		NewInitCliCommand(project, dispatcher),
		NewEnvironmentCliCommand(project, dispatcher),
		NewServiceCliCommand(project, dispatcher),
		NewEnvCliCommand(project, dispatcher),
	}

	return app
}

// CommandFactory is a function that creates an `ecso.Command` from a `cli.Context` and `config.Config`
type CommandFactory func(*cli.Context, *config.Config) (ecso.Command, error)

// MakeEcsoCommandFactory creates and ecso.CommandFactory from our local CommandFactory type
func MakeEcsoCommandFactory(ctx *cli.Context, fn CommandFactory) dispatcher.CommandFactory {
	return dispatcher.CommandFactoryFunc(func(cfg *config.Config) (ecso.Command, error) {
		return fn(ctx, cfg)
	})
}

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(dispatcher dispatcher.Dispatcher, factory CommandFactory, options ...func(*dispatcher.DispatchOptions)) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		if err := dispatcher.Dispatch(MakeEcsoCommandFactory(ctx, factory), options...); err != nil {
			if ecso.IsArgumentRequiredError(err) || ecso.IsOptionRequiredError(err) {
				cli.ShowSubcommandHelp(ctx)
			}

			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	}
}
