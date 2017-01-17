package commands

import "github.com/bernos/ecso/pkg/ecso"

// Command represents a single ecso command
type Command interface {
	Execute(log ecso.Logger) error
}

// CommandFunc lifts a regular function to the Command interface
type CommandFunc func(ecso.Logger) error

// Execute executes the func
func (fn CommandFunc) Execute(log ecso.Logger) error {
	return fn(log)
}

// CommandError wraps an error in a func that satisfies the Command
// interface. Use this to simplify returning errors from functions
// that create commands
func CommandError(err error) Command {
	return CommandFunc(func(log ecso.Logger) error {
		return err
	})
}
