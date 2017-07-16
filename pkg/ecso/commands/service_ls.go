package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

const (
	ServiceLsEnvironmentOption = "environment"
)

func NewServiceLsCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &serviceLsCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type serviceLsCommand struct {
	*EnvironmentCommand
}

func (cmd *serviceLsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	env := cmd.Environment(ctx)

	services, err := cmd.environmentAPI.GetECSServices(env)
	if err != nil {
		return err
	}

	tw := ui.NewTableWriter(w, "|")
	tw.WriteHeader([]byte("SERVICE|ECS SERVICE|TASK|DESIRED|RUNNING|STATUS"))

	for _, s := range services {
		row := fmt.Sprintf(
			"%s|%s|%s|%s|%s|%s",
			localServiceName(s, env, ctx.Project),
			*s.ServiceName,
			util.GetIDFromArn(*s.TaskDefinition),
			*s.DesiredCount,
			*s.RunningCount,
			*s.Status)

		tw.Write([]byte(row))
	}

	_, err = tw.Flush()

	return err
}

func localServiceName(ecsService *ecs.Service, env *ecso.Environment, project *ecso.Project) string {
	for _, s := range project.Services {
		if strings.HasPrefix(*ecsService.ServiceName, s.GetECSTaskDefinitionName(env)+"-Service") {
			return s.Name
		}
	}

	return ""
}
