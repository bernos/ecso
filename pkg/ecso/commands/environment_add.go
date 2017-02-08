package commands

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type EnvironmentAddOptions struct {
	Name                 string
	CloudFormationBucket string
	VPCID                string
	ALBSubnets           string
	InstanceSubnets      string
	Region               string
	Account              string
	InstanceType         string
	Size                 int
	DNSZone              string
	DataDogAPIKey        string
}

func NewEnvironmentAddCommand(environmentName string, options ...func(*EnvironmentAddOptions)) ecso.Command {
	o := &EnvironmentAddOptions{
		Name: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &environmentAddCommand{
		options: o,
	}
}

type environmentAddCommand struct {
	options *EnvironmentAddOptions
}

func (c *environmentAddCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
	)

	if project.HasEnvironment(c.options.Name) {
		return fmt.Errorf("An environment named '%s' already exists for this project.", c.options.Name)
	}

	project.AddEnvironment(&ecso.Environment{
		Name:                 c.options.Name,
		Region:               c.options.Region,
		CloudFormationBucket: c.options.CloudFormationBucket,
		CloudFormationParameters: map[string]string{
			"VPC":             c.options.VPCID,
			"InstanceSubnets": c.options.InstanceSubnets,
			"ALBSubnets":      c.options.ALBSubnets,
			"InstanceType":    c.options.InstanceType,
			"DNSZone":         c.options.DNSZone,
			"ClusterSize":     fmt.Sprintf("%d", c.options.Size),
			"DataDogAPIKey":   c.options.DataDogAPIKey,
		},
		CloudFormationTags: map[string]string{
			"environment": c.options.Name,
			"project":     project.Name,
		},
	})

	if err := project.Save(); err != nil {
		return err
	}

	log.BannerGreen("Successfully added environment '%s' to the project", c.options.Name)

	return nil
}

func (c *environmentAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (c *environmentAddCommand) Prompt(ctx *ecso.CommandContext) error {

	var (
		log             = ctx.Config.Logger()
		project         = ctx.Project
		options         = c.options
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
		Bucket          string
		ALBSubnets      string
		InstanceSubnets string
		InstanceType    string
		Size            string
		DNSZone         string
		DataDogAPIKey   string
	}{
		Name:            "What is the name of your environment?",
		Region:          "Which AWS region will the environment be deployed to?",
		VPC:             "Which VPC would you like to create the environment in?",
		Bucket:          "Which S3 bucket would you like to use to store CloudFormation templates used by ecso?",
		ALBSubnets:      "Which subnets would you like to deploy the load balancer to?",
		InstanceSubnets: "Which subnets would you like to deploy the ECS container instances to?",
		InstanceType:    "What type of instances would you like to add to the ECS cluster?",
		Size:            "How many instances would you like to add to the ECS cluster?",
		DNSZone:         "Which DNS zone would you like to use for service discovery?",
		DataDogAPIKey:   "What is your Data Dog API key?",
	}

	var validators = struct {
		Name            func(string) error
		Region          func(string) error
		VPC             func(string) error
		Bucket          func(string) error
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
		Bucket:          ui.ValidateRequired("Bucket is required"),
		ALBSubnets:      ui.ValidateRequired("ALB subnets are required"),
		InstanceSubnets: ui.ValidateRequired("Instance subnets are required"),
		InstanceType:    ui.ValidateRequired("Instance type is required"),
		DNSZone:         ui.ValidateRequired("DNS zone is required"),
		Size:            ui.ValidateIntBetween(2, 100),
		DataDogAPIKey:   ui.ValidateRequired("DataDog API key is required"),
	}

	// TODO Ask if there is an existing environment?
	// If yes, then ask for the cfn stack id and collect outputs
	log.BannerBlue("Adding a new environment to the %s project", project.Name)

	if account := getCurrentAWSAccount(stsAPI); options.Account == "" {
		options.Account = account
	}

	if ac, ok := prefs.AccountDefaults[options.Account]; ok {
		accountDefaults = ac
	}

	if err := ui.AskStringIfEmptyVar(&options.Name, prompts.Name, "dev", validators.Name); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.Region, prompts.Region, "ap-southeast-2", validators.Region); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.VPCID, prompts.VPC, accountDefaults.VPCID, validators.VPC); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.CloudFormationBucket, prompts.Bucket, getDefaultCloudFormationBucket(options.Account, options.Region), validators.Bucket); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.ALBSubnets, prompts.ALBSubnets, accountDefaults.ALBSubnets, validators.ALBSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.InstanceSubnets, prompts.InstanceSubnets, accountDefaults.InstanceSubnets, validators.InstanceSubnets); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.InstanceType, prompts.InstanceType, "t2.large", validators.InstanceType); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(&options.Size, prompts.Size, 4, validators.Size); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.DNSZone, prompts.DNSZone, accountDefaults.DNSZone, validators.DNSZone); err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(&options.DataDogAPIKey, prompts.DataDogAPIKey, accountDefaults.DataDogAPIKey, validators.DataDogAPIKey); err != nil {
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

func getDefaultCloudFormationBucket(account, region string) string {
	if account == "" || region == "" {
		return ""
	}

	return fmt.Sprintf("ecso-%s-%s", region, account)
}

func getCurrentAWSAccount(svc stsiface.STSAPI) string {
	if resp, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{}); err == nil {
		return *resp.Account
	}
	return ""
}
