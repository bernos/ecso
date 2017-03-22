package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
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

func NewServicePsCommand(name string, serviceAPI api.ServiceAPI, log ecso.Logger) ecso.Command {
	return &servicePsCommand{
		ServiceCommand: &ServiceCommand{
			name: name,
		},
		serviceAPI: serviceAPI,
		log:        log,
	}
}

type servicePsCommand struct {
	*ServiceCommand

	serviceAPI api.ServiceAPI
	log        ecso.Logger
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		service  = ctx.Project.Services[cmd.name]
		env      = ctx.Project.Environments[cmd.environment]
		rows     = make([]*row, 0)
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsAPI   = registry.ECSAPI()
	)

	runningService, err := cmd.serviceAPI.GetECSService(ctx.Project, env, service)

	if err != nil || runningService == nil {
		return err
	}

	tasks, err := ecsAPI.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(env.GetClusterName()),
		ServiceName: runningService.ServiceName,
	})

	if err != nil {
		return err
	}

	resp, err := ecsAPI.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(env.GetClusterName()),
		Tasks:   tasks.TaskArns,
	})

	if err != nil {
		return err
	}

	for _, task := range resp.Tasks {
		newRows, err := rowsFromTask(task, ecsAPI)

		if err != nil {
			return err
		}

		rows = append(rows, newRows...)
	}

	log.Printf("\n")
	printRows(rows, cmd.log)
	log.Printf("\n")

	return nil
}

func rowsFromTask(task *ecs.Task, ecsAPI ecsiface.ECSAPI) ([]*row, error) {
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

		image, err := getContainerImage(*task.TaskDefinitionArn, *c.Name, ecsAPI)

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

func printRows(rows []*row, log ecso.Logger) {
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

func getContainerImage(taskDefinitionArn, containerName string, ecsAPI ecsiface.ECSAPI) (string, error) {
	resp, err := ecsAPI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionArn),
	})

	if err != nil {
		return "", err
	}

	for _, c := range resp.TaskDefinition.ContainerDefinitions {
		if *c.Name == containerName {
			return *c.Image, nil
		}
	}

	return "", nil
}

func getIDFromArn(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
