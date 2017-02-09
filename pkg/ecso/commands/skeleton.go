package commands

import "github.com/bernos/ecso/pkg/ecso"

type SkeletonOptions struct {
	EnvironmentName string
}

func NewSkeletonCommand(environmentName string, options ...func(*SkeletonOptions)) ecso.Command {
	o := &SkeletonOptions{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &skeletonCommand{
		options: o,
	}
}

type skeletonCommand struct {
	options *SkeletonOptions
}

func (cmd *skeletonCommand) Execute(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *skeletonCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *skeletonCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
