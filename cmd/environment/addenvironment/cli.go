package addenvironment

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name                 string
	VPCID                string
	CloudFormationBucket string
	ALBSubnets           string
	InstanceSubnets      string
	InstanceType         string
	Region               string
	Size                 string
}{
	Name:            "name",
	VPCID:           "vpc",
	ALBSubnets:      "alb-subnets",
	InstanceSubnets: "instance-subnets",
	InstanceType:    "instance-type",
	Region:          "region",
	Size:            "size",
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	return commands.NewEnvironmentAddCommand(c.Args().First(), func(opt *commands.EnvironmentAddOptions) {
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

		if c.Int(keys.Size) != 0 {
			opt.Size = c.Int(keys.Size)
		}

		if c.String(keys.InstanceType) != "" {
			opt.InstanceType = c.String(keys.InstanceType)
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
			cli.IntFlag{
				Name:  keys.Size,
				Usage: "Then number of container instances to create",
			},
			cli.StringFlag{
				Name:  keys.InstanceType,
				Usage: "The type of container instances to create",
			},
		},
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}