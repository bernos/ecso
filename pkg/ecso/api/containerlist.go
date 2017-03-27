package api

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
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

func (cs ContainerList) TableHeader() []string {
	return []string{
		"CONTAINER",
		"IMAGE",
		"GROUP",
		"STATUS",
		"TASK NAME",
		"CONTAINER INSTANCE",
		"PORT",
	}
}

func (cs ContainerList) TableRows() []map[string]string {
	trs := make([]map[string]string, len(cs))

	for i, c := range cs {
		trs[i] = map[string]string{
			"CONTAINER":          *c.containerDefinition.Name,
			"IMAGE":              *c.containerDefinition.Image,
			"GROUP":              *c.task.Group,
			"STATUS":             *c.container.LastStatus,
			"TASK NAME":          util.GetIDFromArn(*c.task.TaskDefinitionArn),
			"CONTAINER INSTANCE": util.GetIDFromArn(*c.task.ContainerInstanceArn),
			"PORT":               "",
		}

		if len(c.container.NetworkBindings) > 0 {
			ports := make([]string, 0)

			for _, b := range c.container.NetworkBindings {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", *b.ContainerPort, *b.HostPort, *b.Protocol))
			}

			trs[i]["PORT"] = strings.Join(ports, ",")
		}
	}

	return trs
}
