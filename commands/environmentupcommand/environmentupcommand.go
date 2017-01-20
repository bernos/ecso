package environmentupcommand

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/util"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Unset string
}{
	Unset: "unset",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "environment-up",
		Usage:     "Bring up the named environment",
		ArgsUsage: "[environment]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Unset,
				Usage: "If set, output shell commands to unset all ecso environment variables",
			},
		},
		Action: commands.MakeAction(FromCliContext, dispatcher),
	}
}

func FromCliContext(c *cli.Context) ecso.Command {
	return New(c.Args().First(), func(opt *Options) {

	})
}

type Options struct {
	EnvironmentName string
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &envUpCommand{
		options: o,
	}
}

type envUpCommand struct {
	options *Options
}

func (cmd *envUpCommand) Execute(project *ecso.Project, cfg *ecso.Config, prefs ecso.UserPreferences) error {
	if err := validateOptions(cmd.options); err != nil {
		return err
	}

	_, ok := project.Environments[cmd.options.EnvironmentName]

	if !ok {
		return fmt.Errorf("No environment named '%s' was found", cmd.options.EnvironmentName)
	}

	// Check whether env cfn templates already exist?
	templateDir, err := getTemplateDir()

	if err != nil {
		return err
	}

	exists, err := util.DirExists(templateDir)

	if err != nil {
		return err
	}

	if !exists {
		if err := copyTemplates(templateDir); err != nil {
			return err
		}
	}

	return nil
}

func getTemplateDir() (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(wd, ".ecso", "infrastructure", "templates"), nil
}

func copyTemplates(dst string) error {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	for file, content := range templates {
		if err := ioutil.WriteFile(filepath.Join(dst, file), []byte(content), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func validateOptions(opt *Options) error {
	if opt.EnvironmentName == "" {
		return fmt.Errorf("Environment name is required")
	}
	return nil
}
