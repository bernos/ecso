package commands

import (
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceUpCommand(name string, serviceAPI api.ServiceAPI, log ecso.Logger) ecso.Command {
	return &serviceUpCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceUpCommand struct {
	*ServiceCommand
}

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environment]
		service = project.Services[cmd.name]
	)

	ui.BannerBlue(
		cmd.log,
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	if err := cmd.serviceAPI.ServiceUp(project, env, service); err != nil {
		return err
	}

	description, err := cmd.serviceAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(cmd.log, description)

	ui.BannerGreen(
		cmd.log,
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	return nil
}

func getTemplateDir(serviceName string) (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return wd, err
	}

	return filepath.Join(wd, ".ecso", "helpers", serviceName), nil
}
