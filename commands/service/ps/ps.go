package ps

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type Options struct {
	Name        string
	Environment string
}

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

func New(name, env string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:        name,
		Environment: env,
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
		service  = ctx.Project.Services[cmd.options.Name]
		env      = ctx.Project.Environments[cmd.options.Environment]
		log      = ctx.Config.Logger
		rows     = make([]*row, 0)
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsAPI   = registry.ECSAPI()
	)

	tasks, err := ecsAPI.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(env.GetClusterName()),
		ServiceName: aws.String(service.GetECSServiceName()),
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

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
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

	idWidth := 36
	taskNameWidth := len("TASK NAME")
	desiredStatusWidth := len("DESIRED STATUS")
	currentStatusWidth := len("CURRENT STATUS")
	containerWidth := len("CONTAINER")
	imageWidth := len("IMAGE")
	containerStatusWidth := len("STATUS")

	for _, row := range rows {
		if len(row.TaskName) > taskNameWidth {
			taskNameWidth = len(row.TaskName)
		}

		if len(row.DesiredStatus) > desiredStatusWidth {
			desiredStatusWidth = len(row.DesiredStatus)
		}

		if len(row.CurrentStatus) > currentStatusWidth {
			currentStatusWidth = len(row.CurrentStatus)
		}

		if len(row.ContainerName) > containerWidth {
			containerWidth = len(row.ContainerName)
		}

		if len(row.ImageName) > imageWidth {
			imageWidth = len(row.ImageName)
		}

		if len(row.ContainerStatus) > containerStatusWidth {
			containerStatusWidth = len(row.ContainerStatus)
		}
	}

	headerFormat := fmt.Sprintf(
		"%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n",
		containerWidth,
		imageWidth,
		containerStatusWidth,
		taskNameWidth,
		idWidth,
		desiredStatusWidth,
		currentStatusWidth)

	rowFormat := headerFormat

	log.Printf(
		headerFormat,
		"CONTAINER",
		"IMAGE",
		"STATUS",
		"TASK NAME",
		"CONTAINER INSTANCE",
		"DESIRED STATUS",
		"CURRENT STATUS",
		"PORT")

	for _, row := range rows {
		log.Printf(
			rowFormat,
			row.ContainerName,
			row.ImageName,
			row.ContainerStatus,
			row.TaskName,
			row.ContainerInstance,
			row.DesiredStatus,
			row.CurrentStatus,
			row.Port)
	}
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
