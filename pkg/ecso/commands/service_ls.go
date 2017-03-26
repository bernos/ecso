package commands

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
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

func (cmd *serviceLsCommand) Execute(ctx *ecso.CommandContext) error {
	env := cmd.Environment(ctx)

	services, err := cmd.environmentAPI.GetECSServices(env)

	if err != nil {
		return err
	}

	ui.PrintTable(cmd.log, servicesToRows(ctx.Project, env, services))

	return nil
}

func servicesToRows(project *ecso.Project, env *ecso.Environment, services []*ecs.Service) serviceLsRows {
	a := make([]*serviceLsRow, len(services))

	for i, s := range services {
		a[i] = serviceToRow(project, env, s)
	}

	return serviceLsRows(a)
}

func serviceToRow(project *ecso.Project, env *ecso.Environment, service *ecs.Service) *serviceLsRow {
	return &serviceLsRow{
		Name:           localServiceName(service, env, project),
		ECSServiceName: *service.ServiceName,
		ECSTask:        util.GetIDFromArn(*service.TaskDefinition),
		Desired:        *service.DesiredCount,
		Running:        *service.RunningCount,
		Status:         *service.Status,
	}
}

func localServiceName(ecsService *ecs.Service, env *ecso.Environment, project *ecso.Project) string {
	for _, s := range project.Services {
		if strings.HasPrefix(*ecsService.ServiceName, s.GetECSTaskDefinitionName(env)+"-Service") {
			return s.Name
		}
	}

	return ""
}

type serviceLsRow struct {
	Name           string
	ECSServiceName string
	ECSTask        string
	Desired        int64
	Running        int64
	Status         string
}

type serviceLsRows []*serviceLsRow

func (s serviceLsRows) TableHeader() []string {
	return []string{
		"SERVICE",
		"ECS SERVICE",
		"TASK",
		"DESIRED",
		"RUNNING",
		"STATUS",
	}
}

func (s serviceLsRows) TableRows() []map[string]string {
	trs := make([]map[string]string, len(s))

	for i, row := range s {
		trs[i] = map[string]string{
			"SERVICE":     row.Name,
			"ECS SERVICE": row.ECSServiceName,
			"TASK":        row.ECSTask,
			"DESIRED":     fmt.Sprintf("%d", row.Desired),
			"RUNNING":     fmt.Sprintf("%d", row.Running),
			"STATUS":      row.Status,
		}
	}

	return trs
}
