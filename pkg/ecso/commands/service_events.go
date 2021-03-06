package commands

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceEventsCommand(name string, environmentName string, serviceAPI api.ServiceAPI) *ServiceEventsCommand {
	return &ServiceEventsCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type ServiceEventsCommand struct {
	*ServiceCommand
}

func (cmd *ServiceEventsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		count   = 0
		ew      = ui.NewErrWriter(w)
	)

	cancel, err := cmd.serviceAPI.ServiceEvents(ctx.Project, env, service, func(e *ecs.ServiceEvent, err error) {
		if err != nil {
			fmt.Fprintf(ew, "%s\n", err.Error())
		} else {
			fmt.Fprintf(w, "%s %s\n", *e.CreatedAt, *e.Message)
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
