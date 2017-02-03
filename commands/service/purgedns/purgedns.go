package purgedns

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
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
		env            = ctx.Project.Environments[cmd.options.Environment]
		service        = ctx.Project.Services[cmd.options.Name]
		serviceDNSName = fmt.Sprintf("%s.%s.", service.Name, env.GetClusterName())
	)

	registry, err := ctx.Config.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	svc := registry.Route53API()

	zones, err := svc.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName: aws.String(env.CloudFormationParameters["DNSZone"] + "."),
	})

	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", zones)

	for _, zone := range zones.HostedZones {
		resp, err := svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		})

		if err != nil {
			return err
		}

		for _, record := range resp.ResourceRecordSets {
			fmt.Printf("Considering: %s\n", *record.Name)

			if *record.Name == serviceDNSName {
				fmt.Printf("DELETING...\n")
			}
		}
	}

	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
