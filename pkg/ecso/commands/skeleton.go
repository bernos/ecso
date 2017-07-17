package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
)

func NewSkeletonCommand(environmentName string) *SkeletonCommand {
	return &SkeletonCommand{
		environmentName: environmentName,
	}
}

type SkeletonCommand struct {
	environmentName string
}

func (cmd *SkeletonCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	return nil
}

func (cmd *SkeletonCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
