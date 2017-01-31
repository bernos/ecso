package rm

import (
	"fmt"

	"github.com/bernos/ecso/commands/service/servicedown"
	"github.com/bernos/ecso/pkg/ecso"
)

type Options struct {
	Name string
}

func New(name string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name: name,
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
		log = ctx.Config.Logger
		env = ctx.Project.Environments[cmd.options.Name]
	)

	registry, err := ctx.Config.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	cfnService := registry.CloudFormationService(log.PrefixPrintf("  "))

	log.BannerBlue("Removing '%s' environment", env.Name)

	for _, service := range ctx.Project.Services {
		log.Infof("Removing '%s' service...", service.Name)

		cmd := servicedown.New(service.Name, env.Name)

		if err := cmd.Execute(ctx); err != nil {
			return err
		}
	}

	log.Infof("Deleting Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnService.DeleteStack(env.GetCloudFormationStackName()); err != nil {
		return err
	}

	delete(ctx.Project.Environments, cmd.options.Name)

	if err := ctx.Project.Save(); err != nil {
		return err
	}

	log.BannerGreen("Successfully removed '%s' environment", env.Name)

	return nil
}

func validateOptions(opt *Options, ctx *ecso.CommandContext) error {
	if ctx.Project.Environments[opt.Name] == nil {
		return fmt.Errorf("Environment '%s' not found", opt.Name)
	}

	return nil
}
