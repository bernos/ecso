package addenvironment

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name                 string
	VPCID                string
	CloudFormationBucket string
	ALBSubnets           string
	InstanceSubnets      string
	Region               string
}{
	Name:            "name",
	VPCID:           "vpc",
	ALBSubnets:      "alb-subnets",
	InstanceSubnets: "instance-subnets",
	Region:          "region",
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	return New(c.Args().First(), func(opt *Options) {
		if c.String(keys.CloudFormationBucket) != "" {
			opt.CloudFormationBucket = c.String(keys.CloudFormationBucket)
		}

		if c.String(keys.VPCID) != "" {
			opt.VPCID = c.String(keys.VPCID)
		}

		if c.String(keys.ALBSubnets) != "" {
			opt.ALBSubnets = c.String(keys.ALBSubnets)
		}

		if c.String(keys.InstanceSubnets) != "" {
			opt.InstanceSubnets = c.String(keys.InstanceSubnets)
		}

		if c.String(keys.Region) != "" {
			opt.Region = c.String(keys.Region)
		}
	}), nil
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[environment]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.CloudFormationBucket,
				Usage: "The S3 bucket that ecso will upload cloud formation templates for this environment to. If this bucket does not exist, ecso will create it.",
			},
			cli.StringFlag{
				Name:  keys.VPCID,
				Usage: "The vpc to create the environment in",
			},
			cli.StringFlag{
				Name:  keys.ALBSubnets,
				Usage: "The subnets to place the application load balancer in",
			},
			cli.StringFlag{
				Name:  keys.InstanceSubnets,
				Usage: "The subnets to place the ecs container instances in",
			},
			cli.StringFlag{
				Name:  keys.Region,
				Usage: "The AWS region to create the environment in",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}
