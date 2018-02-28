package cli

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

var ServiceAddFlags = struct {
	DesiredCount cli.IntFlag
	Route        cli.StringFlag
	Port         cli.IntFlag
}{

	DesiredCount: cli.IntFlag{
		Name:  "desired-cout",
		Usage: "The desired number of service instances",
	},
	Route: cli.StringFlag{
		Name:  "route",
		Usage: "If set, the service will be registered with the load balancer at this route",
	},
	Port: cli.IntFlag{
		Name:  "port",
		Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
	},
}

func NewServiceAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return &serviceAddCommandWrapper{ctx, cfg}, nil
	}

	return cli.Command{
		Name:        "add",
		Usage:       "Adds a new service to the project",
		Description: "The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			ServiceAddFlags.DesiredCount,
			ServiceAddFlags.Route,
			ServiceAddFlags.Port,
		},
	}
}

type serviceAddCommandWrapper struct {
	cliCtx *cli.Context
	cfg    *config.Config
}

func (wrapper *serviceAddCommandWrapper) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		blue  = ui.NewBannerWriter(w, ui.BlueBold)
		green = ui.NewBannerWriter(w, ui.GreenBold)

		serviceName  = wrapper.cliCtx.Args().First()
		desiredCount = wrapper.cliCtx.Int(ServiceAddFlags.DesiredCount.Name)
		route        = wrapper.cliCtx.String(ServiceAddFlags.Route.Name)
		port         = wrapper.cliCtx.Int(ServiceAddFlags.Port.Name)
	)

	var prompts = struct {
		Name         string
		DesiredCount string
		Route        string
		Port         string
	}{
		Name:         "What is the name of your service?",
		DesiredCount: "How many instances of the service would you like to run?",
		Route:        "What route would you like to expose the service at?",
		Port:         "Which container port would you like to expose?",
	}

	fmt.Fprintf(blue, "Adding a new service to the %s project", ctx.Project.Name)

	if err := ui.AskStringIfEmptyVar(r, w, &serviceName, prompts.Name, "", serviceNameValidator(ctx.Project)); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(r, w, &desiredCount, prompts.DesiredCount, 1, desiredCountValidator()); err != nil {
		return err
	}

	webChoice, err := ui.Choice(r, w, "Is this a web service?", []string{"Yes", "No"})
	if err != nil {
		return err
	}

	if webChoice == 0 {
		if err := ui.AskStringIfEmptyVar(r, w, &route, prompts.Route, "/"+serviceName, routeValidator()); err != nil {
			return err
		}

		if err := ui.AskIntIfEmptyVar(r, w, &port, prompts.Port, 80, portValidator()); err != nil {
			return err
		}
	}

	cmd := commands.NewServiceAddCommand(serviceName).
		WithDesiredCount(desiredCount).
		WithRoute(route).
		WithPort(port)

	if err := cmd.Execute(ctx, r, w); err != nil {
		return err
	}

	fmt.Fprintf(green, "Service '%s' added successfully.", serviceName)
	fmt.Fprintf(w, "Run `ecso service up %s --environment <ENVIRONMENT>` to deploy.\n\n", serviceName)

	return nil
}

func (cmd *serviceAddCommandWrapper) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func serviceNameValidator(p *ecso.Project) ui.StringValidator {
	return ui.StringValidatorFunc(func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasService(val) {
			return fmt.Errorf("This project already has a service named '%s'. Please choose another name", val)
		}

		return nil
	})
}

func routeValidator() ui.StringValidator {
	return ui.ValidateAny()
}

func desiredCountValidator() ui.IntValidator {
	return ui.ValidateIntBetween(1, 10)
}

func portValidator() ui.IntValidator {
	return ui.ValidateIntBetween(1, 60000)
}
