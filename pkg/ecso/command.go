package ecso

import (
	"github.com/bernos/ecso/pkg/ecso/log"
)

// Command represents a single ecso command
type Command interface {
	Execute(ctx *CommandContext, l log.Logger) error
	Prompt(ctx *CommandContext, l log.Logger) error
	Validate(ctx *CommandContext) error
}

// CommandFunc lifts a regular function to the Command interface
type CommandFunc func(*CommandContext, log.Logger) error

// Execute executes the func
func (fn CommandFunc) Execute(ctx *CommandContext, l log.Logger) error {
	return fn(ctx, l)
}

// Prompt asks for user input
func (fn CommandFunc) Prompt(ctx *CommandContext, l log.Logger) error {
	return nil
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
	return CommandFunc(func(ctx *CommandContext, l log.Logger) error {
		return err
	})
}

// CommandContext provides access to configuration options and preferences
// scoped to a running Command
type CommandContext struct {
	EcsoVersion     string
	Options         CommandOptions
	Project         *Project
	UserPreferences *UserPreferences
}

// NewCommandContext creates a CommandContext
func NewCommandContext(project *Project, preferences *UserPreferences, version string, options CommandOptions) *CommandContext {
	return &CommandContext{
		Options:         options,
		Project:         project,
		UserPreferences: preferences,
		EcsoVersion:     version,
	}
}

// CommandOptions are optional settings used to alter command execution behaviour
type CommandOptions interface {
	String(name string) string
	Bool(name string) bool
	Int(name string) int
}
