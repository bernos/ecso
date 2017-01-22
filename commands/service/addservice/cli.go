package addservice

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name         string
	DesiredCount string
	Route        string
	Port         string
}{
	Name:         "name",
	DesiredCount: "desired-count",
	Route:        "route",
	Port:         "port",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "add",
		Usage: "Adds a new service to the project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "The name of the service",
			},
			cli.IntFlag{
				Name:  keys.DesiredCount,
				Usage: "The desired number of service instances",
			},
			cli.StringFlag{
				Name:  keys.Route,
				Usage: "If set, the service will be registered with the load balancer at this route",
			},
			cli.IntFlag{
				Name:  keys.Port,
				Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
			},
		},
		Action: commands.MakeAction(FromCliContext, dispatcher),
	}
}

func FromCliContext(c *cli.Context) ecso.Command {
	return New(c.String(keys.Name), func(opt *Options) {
		opt.DesiredCount = c.Int(keys.DesiredCount)
		opt.Route = c.String(keys.Route)
		opt.Port = c.Int(keys.Port)
	})
}
