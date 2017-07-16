package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
)

func NewSkeletonCommand(environmentName string) ecso.Command {
	return &skeletonCommand{
		environmentName: environmentName,
	}
}

type skeletonCommand struct {
	environmentName string
}

func (cmd *skeletonCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	return nil
}

func (cmd *skeletonCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
