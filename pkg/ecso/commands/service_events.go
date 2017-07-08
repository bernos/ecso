package commands

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

func NewServiceEventsCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceEventsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceEventsCommand struct {
	*ServiceCommand
}

func (cmd *serviceEventsCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		count   = 0
	)

	cancel, err := cmd.serviceAPI.ServiceEvents(ctx.Project, env, service, func(e *ecs.ServiceEvent, err error) {
		if err != nil {
			l.Errorf("%s\n", err.Error())
		} else {
			l.Printf("%s %s\n", *e.CreatedAt, *e.Message)
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
