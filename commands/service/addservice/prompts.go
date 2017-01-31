package addservice

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func promptForMissingOptions(opt *Options, ctx *ecso.CommandContext) error {
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

	if err := ui.AskStringIfEmptyVar(&opt.Name, prompts.Name, "", serviceNameValidator(ctx.Project)); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(&opt.DesiredCount, prompts.DesiredCount, 1, desiredCountValidator()); err != nil {
		return err
	}

	webChoice, err := ui.Choice("Is this a web service?", []string{"Yes", "No"})

	if err != nil {
		return err
	}

	if webChoice == 0 {
		if err := ui.AskStringIfEmptyVar(&opt.Route, prompts.Route, "/"+opt.Name, routeValidator()); err != nil {
			return err
		}

		if err := ui.AskIntIfEmptyVar(&opt.Port, prompts.Port, 80, portValidator()); err != nil {
			return err
		}
	}

	return nil
}

func serviceNameValidator(p *ecso.Project) func(string) error {
	return func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasService(val) {
			return fmt.Errorf("This project already has a service named '%s'. Please choose another name", val)
		}

		return nil
	}
}

func routeValidator() func(string) error {
	return ui.ValidateAny()
}

func desiredCountValidator() func(int) error {
	return ui.ValidateIntBetween(1, 10)
}

func portValidator() func(int) error {
	return ui.ValidateIntBetween(1, 60000)
}
