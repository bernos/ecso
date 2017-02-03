package purgedns

import (
	"fmt"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name        string
	Environment string
}{
	Name:        "name",
	Environment: "environment",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "purge-dns",
		Usage: "remove stale dns entries for the service",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "the name of the service",
			},
			cli.StringFlag{
				Name:  keys.Environment,
				Usage: "the name of the environment",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	service := c.String(keys.Name)
	env := c.String(keys.Environment)

	if service == "" {
		return nil, commands.NewOptionRequiredError(keys.Name)
	}

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(c.String(keys.Name), c.String(keys.Environment), func(opt *Options) {
		// TODO: populate options from c
	}), nil
}

type Options struct {
	Name        string
	Environment string
}

func New(name, environment string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &command{
		options: o,
	}
}

type command struct {
	options *Options
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	var (
		log            = ctx.Config.Logger
		env            = ctx.Project.Environments[cmd.options.Environment]
		service        = ctx.Project.Services[cmd.options.Name]
		serviceDNSName = fmt.Sprintf("%s.%s.", service.Name, env.GetClusterName())
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
	)

	registry, err := ctx.Config.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	svc := registry.Route53Service(log.PrefixPrintf("  "))

	return svc.DeleteResourceRecordSetsByName(serviceDNSName, zone, "Deleted by ecso service purge-dns")

}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
