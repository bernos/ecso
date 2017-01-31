package addenvironment

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func promptForMissingOptions(options *Options, ctx *ecso.CommandContext) error {
	var (
		cfg             = ctx.Config
		prefs           = ctx.UserPreferences
		accountDefaults = ecso.AccountDefaults{}
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
	}{
		Name:            "What is the name of your environment?",
		Region:          "Which AWS region will the environment be deployed to?",
		VPC:             "Which VPC would you like to create the environment in?",
		Bucket:          "Which S3 bucket would you like to use to store CloudFormation templates used by ecso?",
		ALBSubnets:      "Which subnets would you like to deploy the load balancer to?",
		InstanceSubnets: "Which subnets would you like to deploy the ECS container instances to?",
		InstanceType:    "What type of instances would you like to add to the ECS cluster?",
		Size:            "How many instances would you like to add to the ECS cluster?",
	}

	var validators = struct {
		Name            func(string) error
		Region          func(string) error
		VPC             func(string) error
		Bucket          func(string) error
		ALBSubnets      func(string) error
		InstanceSubnets func(string) error
		InstanceType    func(string) error
		Size            func(int) error
	}{
		Name:            environmentNameValidator(ctx.Project),
		Region:          ui.ValidateRequired("Region is required"),
		VPC:             ui.ValidateRequired("VPC is required"),
		Bucket:          ui.ValidateRequired("Bucket is required"),
		ALBSubnets:      ui.ValidateRequired("ALB subnets are required"),
		InstanceSubnets: ui.ValidateRequired("Instance subnets are required"),
		InstanceType:    ui.ValidateRequired("Instance type is required"),
		Size:            ui.ValidateIntBetween(2, 100),
	}

	// TODO Ask if there is an existing environment?
	// If yes, then ask for the cfn stack id and collect outputs

	registry, err := cfg.GetAWSClientRegistry("ap-southeast-2")

	if err != nil {
		return err
	}

	stsAPI := registry.STSAPI()

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
