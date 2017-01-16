package initcommand

import (
	"fmt"

	"github.com/bernos/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

func FromCliContext(c *cli.Context) commands.Command {
	project := c.Args().First()

	if len(project) == 0 {
		return commands.CommandError(fmt.Errorf("Project is a required parameter."))
	}

	return &initCommand{project}
}

type initCommand struct {
	Project string
}

func (cmd *initCommand) Execute() error {
	fmt.Printf("Initialising project %s\n", cmd.Project)

	return nil
}
