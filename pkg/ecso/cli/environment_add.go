package cli

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

var EnvironmentAddFlags = struct {
	VPC             cli.StringFlag
	ALBSubnets      cli.StringFlag
	InstanceSubnets cli.StringFlag
	InstanceType    cli.StringFlag
	Region          cli.StringFlag
	Size            cli.IntFlag
	KeyPair         cli.StringFlag
	DataDogAPIKey   cli.StringFlag
	DNSZone         cli.StringFlag
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
	DataDogAPIKey: cli.StringFlag{
		Name:  "datadog-api-key",
		Usage: "The DataDog API key to use when sending metrics to datadog",
	},
	DNSZone: cli.StringFlag{
		Name:  "dns-zone",
		Usage: "The DNS zone to create the cluster dns entry in",
	},
}

func NewEnvironmentAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {

		return &environmentAddCommandWrapper{ctx, cfg}, nil
	}

	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[ENVIRONMENT]",
		Flags: []cli.Flag{
			EnvironmentAddFlags.VPC,
			EnvironmentAddFlags.ALBSubnets,
			EnvironmentAddFlags.InstanceSubnets,
			EnvironmentAddFlags.Region,
			EnvironmentAddFlags.Size,
			EnvironmentAddFlags.InstanceType,
			EnvironmentAddFlags.KeyPair,
			EnvironmentAddFlags.DataDogAPIKey,
			EnvironmentAddFlags.DNSZone,
		},
		Action: MakeAction(dispatcher, fn),
	}
}

type environmentAddCommandWrapper struct {
	cliCtx *cli.Context
	cfg    *config.Config
}

func (wrapper *environmentAddCommandWrapper) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (wrapper *environmentAddCommandWrapper) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project         = ctx.Project
		prefs           = ctx.UserPreferences
		blue            = ui.NewBannerWriter(w, ui.BlueBold)
		accountDefaults = ecso.AccountDefaults{}

		environmentName = wrapper.cliCtx.Args().First()
		region          = wrapper.cliCtx.String(EnvironmentAddFlags.Region.Name)
		albSubnets      = wrapper.cliCtx.String(EnvironmentAddFlags.ALBSubnets.Name)
		instanceSubnets = wrapper.cliCtx.String(EnvironmentAddFlags.InstanceSubnets.Name)
		instanceType    = wrapper.cliCtx.String(EnvironmentAddFlags.InstanceType.Name)
		size            = wrapper.cliCtx.Int(EnvironmentAddFlags.Size.Name)
		vpcID           = wrapper.cliCtx.String(EnvironmentAddFlags.VPC.Name)
		keyPair         = wrapper.cliCtx.String(EnvironmentAddFlags.KeyPair.Name)
		datadogAPIKey   = wrapper.cliCtx.String(EnvironmentAddFlags.DataDogAPIKey.Name)
		dnsZone         = wrapper.cliCtx.String(EnvironmentAddFlags.DNSZone.Name)
	)

	var prompts = struct {
		Name            string
		Region          string
		VPC             string
		ALBSubnets      string
		InstanceSubnets string
		InstanceType    string
		Size            string
		KeyPair         string
		DNSZone         string
		DataDogAPIKey   string
	}{
		Name:            "What is the name of your environment?",
		Region:          "Which AWS region will the environment be deployed to?",
		VPC:             "Which VPC would you like to create the environment in (provide the VPC id)?",
		ALBSubnets:      "Which subnets would you like to deploy the load balancer to (provide a comma separated list of subnet ids)?",
		InstanceSubnets: "Which subnets would you like to deploy the ECS container instances to (provide a comma separated list of subnet ids)?",
		InstanceType:    "What type of instances would you like to add to the ECS cluster?",
		Size:            "How many instances would you like to add to the ECS cluster?",
		KeyPair:         "Which keypair would you like to use to access the EC2 isntances in the cluster?",
		DNSZone:         "Which DNS zone would you like to use for service discovery?",
		DataDogAPIKey:   "What is your Data Dog API key?",
	}

	var validators = struct {
		Name            ui.StringValidator
		Region          ui.StringValidator
		VPC             ui.StringValidator
		ALBSubnets      ui.StringValidator
		InstanceSubnets ui.StringValidator
		InstanceType    ui.StringValidator
		DNSZone         ui.StringValidator
		Size            ui.IntValidator
		KeyPair         ui.StringValidator
		DataDogAPIKey   ui.StringValidator
	}{
		Name:            environmentNameValidator(ctx.Project),
		Region:          ui.ValidateRequired("Region is required"),
		VPC:             ui.ValidateRequired("VPC is required"),
		ALBSubnets:      ui.ValidateRequired("ALB subnets are required"),
		InstanceSubnets: ui.ValidateRequired("Instance subnets are required"),
		InstanceType:    ui.ValidateRequired("Instance type is required"),
		DNSZone:         ui.ValidateRequired("DNS zone is required"),
		Size:            ui.ValidateIntBetween(2, 100),
		KeyPair:         ui.ValidateRequired("KeyPair is required"),
		DataDogAPIKey:   ui.ValidateRequired("DataDog API key is required"),
	}

	// TODO Ask if there is an existing environment?
	// If yes, then ask for the cfn stack id and collect outputs
	fmt.Fprintf(blue, "Adding a new environment to the %s project", project.Name)

	if err := ui.AskStringIfEmptyVar(r, w, &environmentName, prompts.Name, "dev", validators.Name); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &region, prompts.Region, "ap-southeast-2", validators.Region); err != nil {
		return err
	}

	environmentAPI := wrapper.cfg.EnvironmentAPI(region)

	if account, err := environmentAPI.GetCurrentAWSAccount(); err == nil {
		if ac, ok := prefs.AccountDefaults[account]; ok {
			accountDefaults = ac
		}
	}

	if err := ui.AskStringIfEmptyVar(r, w, &vpcID, prompts.VPC, accountDefaults.VPCID, validators.VPC); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &albSubnets, prompts.ALBSubnets, accountDefaults.ALBSubnets, validators.ALBSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &instanceSubnets, prompts.InstanceSubnets, accountDefaults.InstanceSubnets, validators.InstanceSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &instanceType, prompts.InstanceType, "t2.large", validators.InstanceType); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(r, w, &size, prompts.Size, 4, validators.Size); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &keyPair, prompts.KeyPair, accountDefaults.KeyPair, validators.KeyPair); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &dnsZone, prompts.DNSZone, accountDefaults.DNSZone, validators.DNSZone); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &datadogAPIKey, prompts.DataDogAPIKey, accountDefaults.DataDogAPIKey, validators.DataDogAPIKey); err != nil {
		return err
	}

	cmd := commands.NewEnvironmentAddCommand(environmentName, environmentAPI).
		WithALBSubnets(albSubnets).
		WithInstanceSubnets(instanceSubnets).
		WithInstanceType(instanceType).
		WithRegion(region).
		WithSize(size).
		WithVPCID(vpcID).
		WithKeyPair(keyPair).
		WithDatadogAPIKey(datadogAPIKey).
		WithDNSZone(dnsZone)

	return cmd.Execute(ctx, r, w)
}

func environmentNameValidator(p *ecso.Project) ui.StringValidator {
	return ui.StringValidatorFunc(func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasEnvironment(val) {
			return fmt.Errorf("This project already contains an environment named '%s', please choose another name", val)
		}
		return nil
	})
}
