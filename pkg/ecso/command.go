package ecso

// Command represents a single ecso command
type Command interface {
	Execute(*Project, *Config, UserPreferences) error
}

// CommandFunc lifts a regular function to the Command interface
type CommandFunc func(*Project, *Config, UserPreferences) error

// Execute executes the func
func (fn CommandFunc) Execute(project *Project, cfg *Config, prefs UserPreferences) error {
	return fn(project, cfg, prefs)
}

// CommandError wraps an error in a func that satisfies the Command
// interface. Use this to simplify returning errors from functions
// that create commands
func CommandError(err error) Command {
	return CommandFunc(func(project *Project, cfg *Config, prefs UserPreferences) error {
		return err
	})
}
