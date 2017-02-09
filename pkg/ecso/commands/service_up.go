package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/services"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceUpOptions struct {
	Name        string
	Environment string
}

func NewServiceUpCommand(name, environment string, options ...func(*ServiceUpOptions)) ecso.Command {
	o := &ServiceUpOptions{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &serviceUpCommand{
		options: o,
	}
}

type serviceUpCommand struct {
	options *ServiceUpOptions
}

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		cfg     = ctx.Config
		log     = cfg.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.options.Environment]
		service = project.Services[cmd.options.Name]
		ecsoAPI = api.New(cfg)
	)

	log.BannerBlue(
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	if err := ecsoAPI.ServiceUp(project, env, service); err != nil {
		return err
	}

	log.BannerGreen(
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	return logOutputs(ctx, env, service)
}

func (cmd *serviceUpCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceUpCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	err := util.AnyError(
		ui.ValidateRequired("Name")(opt.Name),
		ui.ValidateRequired("Environment")(opt.Environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[opt.Name]; !ok {
		return fmt.Errorf("Service '%s' not found", opt.Name)
	}

	if _, ok := ctx.Project.Environments[opt.Environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
	}

	return nil
}

func logOutputs(ctx *ecso.CommandContext, env *ecso.Environment, service *ecso.Service) error {
	var (
		log      = ctx.Config.Logger()
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		cfn      = services.NewCloudFormationService(env.Region, registry.CloudFormationAPI(), registry.S3API(), log.PrefixPrintf("  "))
	)

	outputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return err
	}

	serviceOutputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	items := map[string]string{
		"Service Console": util.ServiceConsoleURL(serviceOutputs["Service"], env.GetClusterName(), env.Region),
	}

	if service.Route != "" {
		items["Service URL"] = fmt.Sprintf("http://%s%s", outputs["RecordSet"], service.Route)
	}

	log.Dl(items)
	log.Printf("\n")

	return nil
}

func getTemplateDir(serviceName string) (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return wd, err
	}

	return filepath.Join(wd, ".ecso", "services", serviceName), nil
}
