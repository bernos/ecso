package commands

import "github.com/bernos/ecso/pkg/ecso"

func NewSkeletonCommand(environmentName string) ecso.Command {
	return &skeletonCommand{
		environmentName: environmentName,
	}
}

type skeletonCommand struct {
	environmentName string
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
