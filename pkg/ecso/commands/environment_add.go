package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	EnvironmentAddVPCOption             = "vpc"
	EnvironmentAddALBSubnetsOption      = "alb-subnets"
	EnvironmentAddInstanceSubnetsOption = "instance-subnets"
	EnvironmentAddInstanceTypeOption    = "instance-type"
	EnvironmentAddRegionOption          = "region"
	EnvironmentAddSizeOption            = "size"
	EnvironmentAddKeyPairOption         = "keypair"
)

type EnvironmentAddCommand struct {
	*EnvironmentCommand

	vpcID           string
	albSubnets      string
	instanceSubnets string
	region          string
	account         string
	instanceType    string
	size            int
	keyPair         string
	dnsZone         string
	datadogAPIKey   string
}

func (c *EnvironmentAddCommand) WithDatadogAPIKey(apiKey string) *EnvironmentAddCommand {
	c.datadogAPIKey = apiKey
	return c
}

func (c *EnvironmentAddCommand) WithDNSZone(zone string) *EnvironmentAddCommand {
	c.dnsZone = zone
	return c
}

func (c *EnvironmentAddCommand) WithKeyPair(keyPair string) *EnvironmentAddCommand {
	c.keyPair = keyPair
	return c
}

func (c *EnvironmentAddCommand) WithSize(size int) *EnvironmentAddCommand {
	c.size = size
	return c
}

func (c *EnvironmentAddCommand) WithInstanceType(instanceType string) *EnvironmentAddCommand {
	c.instanceType = instanceType
	return c
}

func (c *EnvironmentAddCommand) WithAccount(account string) *EnvironmentAddCommand {
	c.account = account
	return c
}

func (c *EnvironmentAddCommand) WithRegion(region string) *EnvironmentAddCommand {
	c.region = region
	return c
}

func (c *EnvironmentAddCommand) WithInstanceSubnets(subnets string) *EnvironmentAddCommand {
	c.instanceSubnets = subnets
	return c
}

func (c *EnvironmentAddCommand) WithALBSubnets(subnets string) *EnvironmentAddCommand {
	c.albSubnets = subnets
	return c
}

func (c *EnvironmentAddCommand) WithVPCID(vpcID string) *EnvironmentAddCommand {
	c.vpcID = vpcID
	return c
}

func NewEnvironmentAddCommand(environmentName string, environmentAPI api.EnvironmentAPI) *EnvironmentAddCommand {
	return &EnvironmentAddCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

func (c *EnvironmentAddCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	project := ctx.Project
	green := ui.NewBannerWriter(w, ui.GreenBold)

	if err := c.prompt(ctx, r, w); err != nil {
		return err
	}

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
			"KeyPair":         c.keyPair,
		},
		CloudFormationTags: map[string]string{
			"environment": c.environmentName,
			"project":     project.Name,
		},
	})

	if err := project.Save(); err != nil {
		return err
	}

	fmt.Fprintf(green, "Successfully added environment '%s' to the project", c.environmentName)
	fmt.Fprintf(w, "Now run `ecso environment up %s` to provision the environment in AWS\n\n", c.environmentName)

	return nil
}

func (c *EnvironmentAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (c *EnvironmentAddCommand) prompt(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project         = ctx.Project
		prefs           = ctx.UserPreferences
		accountDefaults = ecso.AccountDefaults{}
		blue            = ui.NewBannerWriter(w, ui.BlueBold)
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

	if account, _ := c.environmentAPI.GetCurrentAWSAccount(); c.account == "" {
		c.account = account
	}

	if ac, ok := prefs.AccountDefaults[c.account]; ok {
		accountDefaults = ac
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.environmentName, prompts.Name, "dev", validators.Name); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.region, prompts.Region, "ap-southeast-2", validators.Region); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.vpcID, prompts.VPC, accountDefaults.VPCID, validators.VPC); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.albSubnets, prompts.ALBSubnets, accountDefaults.ALBSubnets, validators.ALBSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.instanceSubnets, prompts.InstanceSubnets, accountDefaults.InstanceSubnets, validators.InstanceSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.instanceType, prompts.InstanceType, "t2.large", validators.InstanceType); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(r, w, &c.size, prompts.Size, 4, validators.Size); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.keyPair, prompts.KeyPair, accountDefaults.KeyPair, validators.KeyPair); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.dnsZone, prompts.DNSZone, accountDefaults.DNSZone, validators.DNSZone); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(r, w, &c.datadogAPIKey, prompts.DataDogAPIKey, accountDefaults.DataDogAPIKey, validators.DataDogAPIKey); err != nil {
		return err
	}

	return nil
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
