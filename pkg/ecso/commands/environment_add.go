package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"

	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentAddVPCOption             = "vpc"
	EnvironmentAddALBSubnetsOption      = "alb-subnets"
	EnvironmentAddInstanceSubnetsOption = "instance-subnets"
	EnvironmentAddInstanceTypeOption    = "instance-type"
	EnvironmentAddRegionOption          = "region"
	EnvironmentAddSizeOption            = "size"
)

type environmentAddCommand struct {
	// name            string
	*EnvironmentCommand

	vpcID           string
	albSubnets      string
	instanceSubnets string
	region          string
	account         string
	instanceType    string
	size            int
	dnsZone         string
	datadogAPIKey   string
}

func NewEnvironmentAddCommand(environmentName string) ecso.Command {
	return &environmentAddCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
		},
	}
}

func (c *environmentAddCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := c.EnvironmentCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	c.albSubnets = ctx.String(EnvironmentAddALBSubnetsOption)
	c.instanceSubnets = ctx.String(EnvironmentAddInstanceSubnetsOption)
	c.instanceType = ctx.String(EnvironmentAddInstanceTypeOption)
	c.region = ctx.String(EnvironmentAddRegionOption)
	c.size = ctx.Int(EnvironmentAddSizeOption)
	c.vpcID = ctx.String(EnvironmentAddVPCOption)

	return nil
}

func (c *environmentAddCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
	)

	if project.HasEnvironment(c.environmentName) {
		return fmt.Errorf("An environment named '%s' already exists for this project.", c.environmentName)
	}

	project.AddEnvironment(&ecso.Environment{
		Name:   c.environmentName,
		Region: c.region,
		CloudFormationParameters: map[string]string{
			"VPC":             c.vpcID,
			"InstanceSubnets": c.instanceSubnets,
			"ALBSubnets":      c.albSubnets,
			"InstanceType":    c.instanceType,
			"DNSZone":         c.dnsZone,
			"ClusterSize":     fmt.Sprintf("%d", c.size),
			"DataDogAPIKey":   c.datadogAPIKey,
		},
		CloudFormationTags: map[string]string{
			"environment": c.environmentName,
			"project":     project.Name,
		},
	})

	if err := project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(log, "Successfully added environment '%s' to the project", c.environmentName)
	log.Printf("Now run `ecso environment up %s` to provision the environment in AWS\n\n", c.environmentName)

	return nil
}

func (c *environmentAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (c *environmentAddCommand) Prompt(ctx *ecso.CommandContext) error {

	var (
		log             = ctx.Config.Logger()
		project         = ctx.Project
		cfg             = ctx.Config
		prefs           = ctx.UserPreferences
		accountDefaults = ecso.AccountDefaults{}
		registry        = cfg.MustGetAWSClientRegistry("ap-southeast-2")
		stsAPI          = registry.STSAPI()
	)

	var prompts = struct {
		Name            string
		Region          string
		VPC             string
		ALBSubnets      string
		InstanceSubnets string
		InstanceType    string
		Size            string
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
		DNSZone:         "Which DNS zone would you like to use for service discovery?",
		DataDogAPIKey:   "What is your Data Dog API key?",
	}

	var validators = struct {
		Name            func(string) error
		Region          func(string) error
		VPC             func(string) error
		ALBSubnets      func(string) error
		InstanceSubnets func(string) error
		InstanceType    func(string) error
		DNSZone         func(string) error
		Size            func(int) error
		DataDogAPIKey   func(string) error
	}{
		Name:            environmentNameValidator(ctx.Project),
		Region:          ui.ValidateRequired("Region is required"),
		VPC:             ui.ValidateRequired("VPC is required"),
		ALBSubnets:      ui.ValidateRequired("ALB subnets are required"),
		InstanceSubnets: ui.ValidateRequired("Instance subnets are required"),
		InstanceType:    ui.ValidateRequired("Instance type is required"),
		DNSZone:         ui.ValidateRequired("DNS zone is required"),
		Size:            ui.ValidateIntBetween(2, 100),
		DataDogAPIKey:   ui.ValidateRequired("DataDog API key is required"),
	}

	// TODO Ask if there is an existing environment?
	// If yes, then ask for the cfn stack id and collect outputs
	ui.BannerBlue(log, "Adding a new environment to the %s project", project.Name)

	if account := getCurrentAWSAccount(stsAPI); c.account == "" {
		c.account = account
	}

	if ac, ok := prefs.AccountDefaults[c.account]; ok {
		accountDefaults = ac
	}

	if err := ui.AskStringIfEmptyVar(&c.environmentName, prompts.Name, "dev", validators.Name); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.region, prompts.Region, "ap-southeast-2", validators.Region); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.vpcID, prompts.VPC, accountDefaults.VPCID, validators.VPC); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.albSubnets, prompts.ALBSubnets, accountDefaults.ALBSubnets, validators.ALBSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.instanceSubnets, prompts.InstanceSubnets, accountDefaults.InstanceSubnets, validators.InstanceSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.instanceType, prompts.InstanceType, "t2.large", validators.InstanceType); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(&c.size, prompts.Size, 4, validators.Size); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.dnsZone, prompts.DNSZone, accountDefaults.DNSZone, validators.DNSZone); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&c.datadogAPIKey, prompts.DataDogAPIKey, accountDefaults.DataDogAPIKey, validators.DataDogAPIKey); err != nil {
		return err
	}

	return nil
}

func environmentNameValidator(p *ecso.Project) func(string) error {
	return func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasEnvironment(val) {
			return fmt.Errorf("This project already contains an environment named '%s', please choose another name", val)
		}
		return nil
	}
}

func getCurrentAWSAccount(svc stsiface.STSAPI) string {
	if resp, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{}); err == nil {
		return *resp.Account
	}
	return ""
}
