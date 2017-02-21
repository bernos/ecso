package commands

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
	"gopkg.in/urfave/cli.v1"
)

const ServicePsEnvironmentOption = "environment"

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

func NewServicePsCommand(name string) ecso.Command {
	return &servicePsCommand{
		name: name,
	}
}

type servicePsCommand struct {
	name        string
	environment string
}

func (cmd *servicePsCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServicePsEnvironmentOption)
	return nil
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		service  = ctx.Project.Services[cmd.name]
		env      = ctx.Project.Environments[cmd.environment]
		log      = ctx.Config.Logger()
		rows     = make([]*row, 0)
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsAPI   = registry.ECSAPI()
		ecsoAPI  = api.New(ctx.Config)
	)

	runningService, err := ecsoAPI.GetECSService(ctx.Project, env, service)

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
	printRows(rows, log)
	log.Printf("\n")

	return nil
}

func (cmd *servicePsCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *servicePsCommand) Validate(ctx *ecso.CommandContext) error {
	err := util.AnyError(
		ui.ValidateRequired("Name")(cmd.name),
		ui.ValidateRequired("Environment")(cmd.environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[cmd.name]; !ok {
		return fmt.Errorf("Service '%s' not found", cmd.name)
	}

	if _, ok := ctx.Project.Environments[cmd.environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", cmd.environment)
	}

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
