package ecso

import "gopkg.in/urfave/cli.v1"

// CliContextUnmarshaller is an interface that can unmarshal a
// Context struct from the urfave/cli package
type CliContextUnmarshaller interface {
	UnmarshalCliContext(c *cli.Context) error
}

// Command represents a single ecso command
type Command interface {
	CliContextUnmarshaller
	Prompt(ctx *CommandContext) error
	Validate(ctx *CommandContext) error
	Execute(ctx *CommandContext) error
}

// CommandFunc lifts a regular function to the Command interface
type CommandFunc func(*CommandContext) error

// Execute executes the func
func (fn CommandFunc) Execute(ctx *CommandContext) error {
	return fn(ctx)
}

// Prompt asks for user input
func (fn CommandFunc) Prompt(ctx *CommandContext) error {
	return nil
}

// Validate ensures the command is valid. A CommandFunc is always
// valid, as it has no internal state
func (fn CommandFunc) Validate(ctx *CommandContext) error {
	return nil
}

// UnmarshalCliContext unmarshals a cli.Context struct into a
// Command. For a CommandFunc this does nothing, as there is no
// internal state
func (fn CommandFunc) UnmarshalCliContext(c *cli.Context) error {
	return nil
}

// CommandError wraps an error in a func that satisfies the Command
// interface. Use this to simplify returning errors from functions
// that create commands
func CommandError(err error) Command {
	return CommandFunc(func(ctx *CommandContext) error {
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
}
