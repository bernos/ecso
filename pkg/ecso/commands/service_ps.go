package commands

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
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
		service = ctx.Project.Services[cmd.name]
		env     = ctx.Project.Environments[cmd.environment]
		rows    = make([]*row, 0)
	)

	tasks, err := cmd.serviceAPI.GetECSTasks(ctx.Project, env, service)

	if err != nil {
		return err
	}

	for _, task := range tasks {
		newRows, err := rowsFromTask(env, task, cmd.serviceAPI)

		if err != nil {
			return err
		}

		rows = append(rows, newRows...)
	}

	cmd.log.Printf("\n")
	printRows(rows, cmd.log)
	cmd.log.Printf("\n")

	return nil
}

func rowsFromTask(env *ecso.Environment, task *ecs.Task, serviceAPI api.ServiceAPI) ([]*row, error) {
	rows := make([]*row, 0)

	for _, c := range task.Containers {
		row := &row{
			TaskID:            getIDFromArn(*task.TaskArn),
			TaskName:          getIDFromArn(*task.TaskDefinitionArn),
			ContainerInstance: getIDFromArn(*task.ContainerInstanceArn),
			DesiredStatus:     *task.DesiredStatus,
			CurrentStatus:     *task.LastStatus,
			ContainerName:     *c.Name,
			ContainerStatus:   *c.LastStatus,
		}

		image, err := serviceAPI.GetECSContainerImage(*task.TaskDefinitionArn, *c.Name, env)

		if err != nil {
			return rows, err
		}

		row.ImageName = image

		if len(c.NetworkBindings) > 0 {
			ports := make([]string, 0)

			for _, b := range c.NetworkBindings {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", *b.ContainerPort, *b.HostPort, *b.Protocol))
			}

			row.Port = strings.Join(ports, ",")
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func printRows(rows []*row, log log.Logger) {
	headers := []string{
		"CONTAINER",
		"IMAGE",
		"STATUS",
		"TASK NAME",
		"CONTAINER INSTANCE",
		"DESIRED STATUS",
		"CURRENT STATUS",
		"PORT",
	}

	r := make([]map[string]string, len(rows))

	for i, row := range rows {
		r[i] = map[string]string{
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

	ui.PrintTable(log, headers, r...)
}

func getIDFromArn(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
