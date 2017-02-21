package commands

import (
	"fmt"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

const (
	EnvUnsetOption = "unset"
)

func NewEnvCommand(environmentName string) ecso.Command {
	return &envCommand{
		environmentName: environmentName,
	}
}

type envCommand struct {
	environmentName string
	unset           bool
}

func (cmd *envCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environmentName = ctx.Args().First()
	cmd.unset = ctx.Bool(EnvUnsetOption)

	return nil
}

func (cmd *envCommand) Execute(ctx *ecso.CommandContext) error {
	if cmd.unset {
		oldPS1 := os.Getenv("ECSO_OLD_PS1")

		fmt.Printf("unset ECSO_ENVIRONMENT; ")
		fmt.Printf("unset ECSO_OLD_PS1; ")

		if oldPS1 != "" {
			fmt.Printf("export PS1=\"%s\"\n", oldPS1)
		}
	} else if cmd.environmentName != "" {
		if _, ok := ctx.Project.Environments[cmd.environmentName]; ok {

			ps1 := os.Getenv("PS1")

			if ps1 != "" {
				fmt.Printf("export PS1=\"%s $(tput setaf 2)[ecso::%s:%s]:$(tput sgr0) \"\n", ps1, ctx.Project.Name, cmd.environmentName)
				fmt.Printf("export ECSO_OLD_PS1=\"%s\"\n", ps1)
			}

			fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.environmentName)
		}
	}
	return nil
}

func (cmd *envCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *envCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
