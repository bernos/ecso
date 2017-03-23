package commands

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceLsEnvironmentOption = "environment"
)

func NewServiceLsCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
	return &serviceLsCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
		},
	}
}

type serviceLsCommand struct {
	*EnvironmentCommand
}

func (cmd *serviceLsCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
}

func (cmd *serviceLsCommand) Execute(ctx *ecso.CommandContext) error {
	env := ctx.Project.Environments[cmd.environmentName]

	services, err := cmd.environmentAPI.GetECSServices(env)

	if err != nil {
		return err
	}

	printServices(ctx.Project, env, services, cmd.log)

	return nil
}

func printServices(project *ecso.Project, env *ecso.Environment, services []*ecs.Service, log log.Logger) {
	headers := []string{"SERVICE", "ECS SERVICE", "TASK", "DESIRED", "RUNNING", "STATUS"}
	rows := make([]map[string]string, len(services))

	for i, service := range services {
		rows[i] = map[string]string{
			"SERVICE":     localServiceName(*service.ServiceName, env, project),
			"ECS SERVICE": *service.ServiceName,
			"TASK":        taskDefinitionName(*service.TaskDefinition),
			"DESIRED":     fmt.Sprintf("%d", *service.DesiredCount),
			"RUNNING":     fmt.Sprintf("%d", *service.RunningCount),
			"STATUS":      *service.Status,
		}
	}

	ui.PrintTable(log, headers, rows...)
}

func localServiceName(ecsServiceName string, env *ecso.Environment, project *ecso.Project) string {
	for _, s := range project.Services {
		if strings.HasPrefix(ecsServiceName, s.GetECSTaskDefinitionName(env)+"-Service") {
			return s.Name
		}
	}

	return ""
}

func taskDefinitionName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}

func serviceName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
