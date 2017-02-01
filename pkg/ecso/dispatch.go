package ecso

import "fmt"

// NewDispatcher creates a default Dispatcher for a Project, with the provided Config and
// UserPreferences
func NewDispatcher(project *Project, cfg *Config, prefs *UserPreferences) Dispatcher {
	return DispatcherFunc(func(cmd Command, options ...func(*DispatchOptions)) error {
		opt := &DispatchOptions{
			EnsureProjectExists: true,
		}

		for _, o := range options {
			o(opt)
		}

		if opt.EnsureProjectExists && project == nil {
			return fmt.Errorf("No ecso project file was found")
		}

		ctx := NewCommandContext(project, cfg, prefs)

		if err := cmd.Prompt(ctx); err != nil {
			return err
		}

		if err := cmd.Validate(ctx); err != nil {
			return err
		}

		return cmd.Execute(NewCommandContext(project, cfg, prefs))
	})
}

// Dispatcher executes an ecso Command
type Dispatcher interface {
	Dispatch(Command, ...func(*DispatchOptions)) error
}

// DispatcherFunc is an adaptor to allow the use of ordinary functions as
// an ecso Dispatcher
type DispatcherFunc func(Command, ...func(*DispatchOptions)) error

// Dispatch calls fn(cmd, options...)
func (fn DispatcherFunc) Dispatch(cmd Command, options ...func(*DispatchOptions)) error {
	return fn(cmd, options...)
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
