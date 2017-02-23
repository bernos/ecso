package ecso

import (
	"gopkg.in/urfave/cli.v1"
)

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

func (fn CommandFunc) Prompt(ctx *CommandContext) error {
	return nil
}

func (fn CommandFunc) Validate(ctx *CommandContext) error {
	return nil
}

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
	Project         *Project
	Config          *Config
	UserPreferences *UserPreferences
}

// NewCommandContext creates a CommandContext
func NewCommandContext(project *Project, config *Config, preferences *UserPreferences) *CommandContext {
	return &CommandContext{
		Project:         project,
		Config:          config,
		UserPreferences: preferences,
	}
}
