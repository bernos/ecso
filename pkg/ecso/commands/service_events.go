package commands

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

func NewServiceEventsCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
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

func (cmd *serviceEventsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		count   = 0
	)

	cancel, err := cmd.serviceAPI.ServiceEvents(ctx.Project, env, service, func(e *ecs.ServiceEvent, err error) {
		if err != nil {
			cmd.log.Errorf("%s\n", err.Error())
		} else {
			cmd.log.Printf("%s %s\n", *e.CreatedAt, *e.Message)
		}
	})

	if err != nil {
		return err
	}

	defer cancel()

	for count < 10 {
		time.Sleep(time.Second * 60)
		count++
	}

	return nil
}
