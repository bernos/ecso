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

type row struct {
	TaskID            string
	TaskName          string
	ContainerInstance string
	DesiredStatus     string
	CurrentStatus     string
	ContainerName     string
	ImageName         string
	ContainerStatus   string
	Port              string
}

type rows []*row

func (r rows) TableHeader() []string {
	return []string{
		"CONTAINER",
		"IMAGE",
		"STATUS",
		"TASK NAME",
		"CONTAINER INSTANCE",
		"DESIRED STATUS",
		"CURRENT STATUS",
		"PORT",
	}
}

func (r rows) TableRows() []map[string]string {
	trs := make([]map[string]string, len(r))

	for i, row := range r {
		trs[i] = map[string]string{
			"CONTAINER":          row.ContainerName,
			"IMAGE":              row.ImageName,
			"STATUS":             row.ContainerStatus,
			"TASK NAME":          row.TaskName,
			"CONTAINER INSTANCE": row.ContainerInstance,
			"DESIRED STATUS":     row.DesiredStatus,
			"CURRENT STATUS":     row.CurrentStatus,
			"PORT":               row.Port,
		}
	}

	return trs
}

func NewServicePsCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &servicePsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type servicePsCommand struct {
	*ServiceCommand
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	tasks, err := cmd.serviceAPI.GetECSTasks(ctx.Project, env, service)

	if err != nil {
		return err
	}

	rows, err := cmd.rowsFromTasks(tasks, env)

	if err != nil {
		return err
	}

	cmd.log.Printf("\n")
	ui.PrintTable(cmd.log, rows)
	cmd.log.Printf("\n")

	return nil
}

func (cmd *servicePsCommand) rowsFromTasks(tasks []*ecs.Task, env *ecso.Environment) (rows, error) {
	rs := make([]*row, 0)

	for _, task := range tasks {
		newRows, err := cmd.rowsFromTask(env, task)

		if err != nil {
			return nil, err
		}

		rs = append(rs, newRows...)
	}

	return rows(rs), nil
}

func (cmd *servicePsCommand) rowsFromTask(env *ecso.Environment, task *ecs.Task) (rows, error) {
	rs := make([]*row, 0)

	for _, container := range task.Containers {
		row, err := cmd.rowFromContainer(task, container, env)

		if err != nil {
			return rows(rs), err
		}

		rs = append(rs, row)
	}

	return rows(rs), nil
}

func (cmd *servicePsCommand) rowFromContainer(task *ecs.Task, container *ecs.Container, env *ecso.Environment) (*row, error) {
	row := &row{
		TaskID:            util.GetIDFromArn(*task.TaskArn),
		TaskName:          util.GetIDFromArn(*task.TaskDefinitionArn),
		ContainerInstance: util.GetIDFromArn(*task.ContainerInstanceArn),
		DesiredStatus:     *task.DesiredStatus,
		CurrentStatus:     *task.LastStatus,
		ContainerName:     *container.Name,
		ContainerStatus:   *container.LastStatus,
	}

	image, err := cmd.serviceAPI.GetECSContainerImage(*task.TaskDefinitionArn, *container.Name, env)

	if err != nil {
		return nil, err
	}

	row.ImageName = image

	if len(container.NetworkBindings) > 0 {
		ports := make([]string, 0)

		for _, b := range container.NetworkBindings {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", *b.ContainerPort, *b.HostPort, *b.Protocol))
		}

		row.Port = strings.Join(ports, ",")
	}

	return row, nil
}
