package commands

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/helpers"
)

func NewServiceEventsCommand(name string, serviceAPI api.ServiceAPI, log ecso.Logger) ecso.Command {
	return &serviceEventsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceEventsCommand struct {
	*ServiceCommand
}

// TODO add GetServiceEvents to the ecso api and call from here, rather than using the helper directly
func (cmd *serviceEventsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env       = ctx.Project.Environments[cmd.environment]
		service   = ctx.Project.Services[cmd.name]
		registry  = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsHelper = helpers.NewECSHelper(registry.ECSAPI(), cmd.log.Child())
		count     = 0
	)

	runningService, err := cmd.serviceAPI.GetECSService(ctx.Project, env, service)

	if err != nil || runningService == nil {
		return err
	}

	cancel := ecsHelper.LogServiceEvents(*runningService.ServiceName, env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
		if err != nil {
			cmd.log.Errorf("%s\n", err.Error())
		} else {
			cmd.log.Printf("%s %s\n", *e.CreatedAt, *e.Message)
		}
	})

	defer cancel()

	for count < 10 {
		time.Sleep(time.Second * 60)
		count++
	}

	return nil
}
