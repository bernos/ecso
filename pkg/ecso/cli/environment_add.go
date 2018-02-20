package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		VPC             cli.StringFlag
		ALBSubnets      cli.StringFlag
		InstanceSubnets cli.StringFlag
		InstanceType    cli.StringFlag
		Region          cli.StringFlag
		Size            cli.IntFlag
		KeyPair         cli.StringFlag
	}{
		VPC: cli.StringFlag{
			Name:  "vpc",
			Usage: "The vpc to create the environment in",
		},
		ALBSubnets: cli.StringFlag{
			Name:  "alb-subnets",
			Usage: "The subnets to place the application load balancer in",
		},
		InstanceSubnets: cli.StringFlag{
			Name:  "instance-subnets",
			Usage: "The subnets to place the ecs container instances in",
		},
		Region: cli.StringFlag{
			Name:  "region",
			Usage: "The AWS region to create the environment in",
		},
		Size: cli.IntFlag{
			Name:  "size",
			Usage: "Then number of container instances to create",
		},
		InstanceType: cli.StringFlag{
			Name:  "instance-type",
			Usage: "The type of container instances to create",
		},
		KeyPair: cli.StringFlag{
			Name:  "keypair",
			Usage: "The keypair to use when accessing EC2 instances",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		region := ctx.String(flags.Region.Name)

		if region == "" {
			region = "ap-southeast-2"
		}

		return commands.NewEnvironmentAddCommand(ctx.Args().First(), cfg.EnvironmentAPI(region)).
			WithALBSubnets(ctx.String(flags.ALBSubnets.Name)).
			WithInstanceSubnets(ctx.String(flags.InstanceSubnets.Name)).
			WithInstanceType(ctx.String(flags.InstanceType.Name)).
			WithRegion(ctx.String(flags.Region.Name)).
			WithSize(ctx.Int(flags.Size.Name)).
			WithVPCID(ctx.String(flags.VPC.Name)), nil
	}

	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[ENVIRONMENT]",
		Flags: []cli.Flag{
			flags.VPC,
			flags.ALBSubnets,
			flags.InstanceSubnets,
			flags.Region,
			flags.Size,
			flags.InstanceType,
			flags.KeyPair,
		},
		Action: MakeAction(dispatcher, fn),
	}
}
