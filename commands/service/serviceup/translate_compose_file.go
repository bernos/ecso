package serviceup

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
)

func ConvertToTaskDefinition(taskName, dockerComposeFile string) (*ecs.TaskDefinition, error) {

	envLookup, err := utils.GetDefaultEnvironmentLookup()

	if err != nil {
		return nil, err
	}

	resourceLookup, err := utils.GetDefaultResourceLookup()

	if err != nil {
		return nil, err
	}

	context := &ctx.Context{
		Context: project.Context{
			ComposeFiles:      []string{dockerComposeFile},
			ProjectName:       taskName,
			EnvironmentLookup: envLookup,
			ResourceLookup:    resourceLookup,
		},
	}

	p, err := docker.NewProject(context, nil)

	if err != nil {
		return nil, err
	}

	serviceConfigs := p.(*project.Project).ServiceConfigs

	return utils.ConvertToTaskDefinition(taskName, &context.Context, serviceConfigs)
}
