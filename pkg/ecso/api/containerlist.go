package api

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

func LoadContainerList(tasks []*ecs.Task, ecsAPI ecsiface.ECSAPI) (ContainerList, error) {
	result := make([]*Container, 0)

	for _, task := range tasks {

		taskDefinition, err := ecsAPI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
			TaskDefinition: task.TaskDefinitionArn,
		})

		if err != nil {
			return nil, err
		}

		for _, container := range task.Containers {

			for _, containerDefinition := range taskDefinition.TaskDefinition.ContainerDefinitions {
				if *containerDefinition.Name == *container.Name {
					result = append(result, &Container{
						task:                task,
						container:           container,
						containerDefinition: containerDefinition,
					})

					break
				}
			}
		}
	}

	return result, nil
}

type Container struct {
	task                *ecs.Task
	container           *ecs.Container
	containerDefinition *ecs.ContainerDefinition
}

type ContainerList []*Container

func (cs ContainerList) WriteTo(w io.Writer) (int64, error) {
	tw := ui.NewTableWriter(w, "|")

	if n, err := tw.WriteHeader([]byte("CONTAINER|IMAGE|GROUP|STATUS|TASK NAME|CONATINER INSTANCE|PORT")); err != nil {
		return int64(n), err
	}

	for _, c := range cs {
		port := ""

		if len(c.container.NetworkBindings) > 0 {
			ports := make([]string, 0)

			for _, b := range c.container.NetworkBindings {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", *b.ContainerPort, *b.HostPort, *b.Protocol))
			}

			port = strings.Join(ports, ",")
		}

		row := fmt.Sprintf(
			"%s|%s|%s|%s|%s|%s|%s",
			*c.containerDefinition.Name,
			*c.containerDefinition.Image,
			*c.task.Group,
			*c.container.LastStatus,
			util.GetIDFromArn(*c.task.TaskDefinitionArn),
			util.GetIDFromArn(*c.task.ContainerInstanceArn),
			port)

		if n, err := tw.Write([]byte(row)); err != nil {
			return int64(n), err
		}
	}

	n, err := tw.Flush()

	return int64(n), err
}
