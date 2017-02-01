package logs

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	var (
		cfg       = ctx.Config
		service   = ctx.Project.Services[cmd.options.Name]
		env       = ctx.Project.Environments[cmd.options.Environment]
		log       = ctx.Config.Logger
		registry  = cfg.MustGetAWSClientRegistry(env.Region)
		cfn       = registry.CloudFormationService(log.PrefixPrintf("  "))
		cwLogsAPI = registry.CloudWatchLogsAPI()
	)

	outputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	if err != nil {
		return err
	}

	logGroup := outputs["CloudWatchLogsGroup"]

	if logGroup != "" {

		resp, err := cwLogsAPI.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(logGroup),
		})

		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			log.Printf("%-42s %s\n", time.Unix(*event.Timestamp/1000, *event.Timestamp%1000), *event.Message)
		}
	}

	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	if cmd.options.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if cmd.options.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(cmd.options.Name) {
		return fmt.Errorf("No service named '%s' was found", cmd.options.Name)
	}

	if !ctx.Project.HasEnvironment(cmd.options.Environment) {
		return fmt.Errorf("No environment named '%s' was found", cmd.options.Environment)
	}

	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
