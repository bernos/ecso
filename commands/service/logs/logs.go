package logs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/bernos/ecso/pkg/ecso"
)

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
	if err := validateOptions(cmd.options, ctx); err != nil {
		return err
	}

	var (
		cfg     = ctx.Config
		service = ctx.Project.Services[cmd.options.Name]
		env     = ctx.Project.Environments[cmd.options.Environment]
		log     = ctx.Config.Logger
	)

	cfn, err := cfg.CloudFormationService(env.Region)

	if err != nil {
		return err
	}

	outputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	if err != nil {
		return err
	}

	logGroup := outputs["CloudWatchLogsGroup"]

	if logGroup != "" {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(env.Region),
		})

		if err != nil {
			return err
		}

		cfnLogs := cloudwatchlogs.New(sess)

		resp, err := cfnLogs.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(logGroup),
		})

		if err != nil {
			return err
		}

		log.Printf("%#v\n", resp)
	}

	return nil
}

func validateOptions(opt *Options, ctx *ecso.CommandContext) error {
	if opt.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if opt.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(opt.Name) {
		return fmt.Errorf("No service named '%s' was found", opt.Name)
	}

	if !ctx.Project.HasEnvironment(opt.Environment) {
		return fmt.Errorf("No environment named '%s' was found", opt.Environment)
	}

	return nil
}
