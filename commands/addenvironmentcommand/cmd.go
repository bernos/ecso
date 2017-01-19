package addenvironmentcommand

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

var prompts = struct {
	Name            string
	Region          string
	VPC             string
	Bucket          string
	ALBSubnets      string
	InstanceSubnets string
}{
	Name:            "What is the name of your environment?",
	Region:          "Which AWS region will the environment be deployed to?",
	VPC:             "Which VPC would you like to create the environment in?",
	Bucket:          "Which S3 bucket would you like to use to store CloudFormation templates used by ecso?",
	ALBSubnets:      "Which subnets would you like to deploy the load balancer to?",
	InstanceSubnets: "Which subnets would you like to deploy the ECS container instances to?",
}

var validators = struct {
	Name            func(string) error
	Region          func(string) error
	VPC             func(string) error
	Bucket          func(string) error
	ALBSubnets      func(string) error
	InstanceSubnets func(string) error
}{
	Name:            ui.ValidateNotEmpty("Name is required"),
	Region:          ui.ValidateNotEmpty("Region is required"),
	VPC:             ui.ValidateNotEmpty("VPC is required"),
	Bucket:          ui.ValidateNotEmpty("Bucket is required"),
	ALBSubnets:      ui.ValidateNotEmpty("ALB subnets are required"),
	InstanceSubnets: ui.ValidateNotEmpty("Instance subnets are required"),
}

type Options struct {
	Name                 string
	CloudFormationBucket string
	VPCID                string
	ALBSubnets           string
	InstanceSubnets      string
	Region               string
	Account              string
}

func New(environmentName string, options ...func(*Options)) commands.Command {
	o := &Options{
		Name: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &cmd{
		options: o,
	}
}

type cmd struct {
	options *Options
}

func (c *cmd) Execute(cfg *ecso.Config) error {
	log := cfg.Logger

	project, err := util.LoadCurrentProject()

	if err != nil {
		return err
	}

	log.BannerBlue("Adding a new environment to the %s project", project.Name)

	if err := promptForMissingOptions(c.options, project, cfg); err != nil {
		return err
	}

	if _, ok := project.Environments[c.options.Name]; ok {
		return fmt.Errorf("An environment named '%s' already exists for this project.", c.options.Name)
	}

	environment := ecso.Environment{
		Name:                 c.options.Name,
		CloudFormationBucket: c.options.CloudFormationBucket,
		CloudFormationParameters: map[string]string{
			"VPC":             c.options.VPCID,
			"InstanceSubnets": c.options.InstanceSubnets,
			"ALBSubnets":      c.options.ALBSubnets,
		},
	}

	project.AddEnvironment(c.options.Name, environment)

	err = util.SaveCurrentProject(project)

	if err != nil {
		return err
	}

	log.BannerGreen("Successfully added environment '%s' to the project", environment.Name)

	return nil
}

func promptForMissingOptions(options *Options, project *ecso.Project, cfg *ecso.Config) error {
	var (
		accountDefaults = ecso.AccountDefaults{}
	)

	preferences, err := util.LoadUserPreferences()

	if err != nil {
		return err
	}

	// TODO Ask if there is an existing environment?
	// If yes, then ask for the cfn stack id and collect outputs

	if account := getCurrentAWSAccount(cfg.STS); options.Account == "" {
		options.Account = account
	}

	if ac, ok := preferences.AccountDefaults[options.Account]; ok {
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

	return nil
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
