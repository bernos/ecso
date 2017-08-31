package ecso

import "io"

// Command represents a single ecso command
type Command interface {
	Execute(ctx *CommandContext, r io.Reader, w io.Writer) error
	Validate(ctx *CommandContext) error
}

// CommandFunc lifts a regular function to the Command interface
type CommandFunc func(*CommandContext, io.Reader, io.Writer) error

// Execute executes the func
func (fn CommandFunc) Execute(ctx *CommandContext, r io.Reader, w io.Writer) error {
	return fn(ctx, r, w)
}

// Validate ensures the command is valid. A CommandFunc is always
// valid, as it has no internal state
func (fn CommandFunc) Validate(ctx *CommandContext) error {
	return nil
}

// CommandError wraps an error in a func that satisfies the Command
// interface. Use this to simplify returning errors from functions
// that create commands
func CommandError(err error) Command {
	return CommandFunc(func(ctx *CommandContext, r io.Reader, w io.Writer) error {
		return err
	})
}

// CommandContext provides access to configuration and preferences scoped to a running Command
type CommandContext struct {
	EcsoVersion     string
	Project         *Project
	UserPreferences *UserPreferences
}

// NewCommandContext creates a CommandContext
func NewCommandContext(project *Project, preferences *UserPreferences, version string) *CommandContext {
	return &CommandContext{
		Project:         project,
		UserPreferences: preferences,
		EcsoVersion:     version,
	}
}
