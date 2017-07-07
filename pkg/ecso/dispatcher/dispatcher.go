package dispatcher

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/config"
)

type CommandFactory func(*config.Config) (ecso.Command, error)

// NewDispatcher creates a default Dispatcher for a Project, with the provided Config and
// UserPreferences
func NewDispatcher(project *ecso.Project, cfg *config.Config, prefs *ecso.UserPreferences) Dispatcher {
	return DispatcherFunc(func(factory CommandFactory, cOptions ecso.CommandOptions, options ...func(*DispatchOptions)) error {
		opt := &DispatchOptions{
			EnsureProjectExists: true,
		}

		for _, o := range options {
			o(opt)
		}

		if opt.EnsureProjectExists && project == nil {
			return fmt.Errorf("No ecso project file was found")
		}

		ctx := ecso.NewCommandContext(project, prefs, cfg.Version, cOptions)

		cmd, err := factory(cfg)

		if err != nil {
			return err
		}

		if err := cmd.Prompt(ctx, cfg.Logger()); err != nil {
			return err
		}

		if err := cmd.Validate(ctx); err != nil {
			return err
		}

		return cmd.Execute(ctx, cfg.Logger())
	})
}

// Dispatcher executes an ecso Command
type Dispatcher interface {
	Dispatch(CommandFactory, ecso.CommandOptions, ...func(*DispatchOptions)) error
}

// DispatcherFunc is an adaptor to allow the use of ordinary functions as
// an ecso Dispatcher
type DispatcherFunc func(CommandFactory, ecso.CommandOptions, ...func(*DispatchOptions)) error

// Dispatch calls fn(cmd, options...)
func (fn DispatcherFunc) Dispatch(factory CommandFactory, cOptions ecso.CommandOptions, dOptions ...func(*DispatchOptions)) error {
	return fn(factory, cOptions, dOptions...)
}

// DispatchOptions alter how a Dispatcher dispatches Commands
type DispatchOptions struct {
	// EnsureProjectExists determines whether the dispatcher will return an
	// error if the Project it is dispatching the Command on is nil
	EnsureProjectExists bool
}

// SkipEnsureProjectExists is an option function that will permit dispatching
// a Command on a Project that is nil
func SkipEnsureProjectExists() func(*DispatchOptions) {
	return func(opt *DispatchOptions) {
		opt.EnsureProjectExists = false
	}
}
