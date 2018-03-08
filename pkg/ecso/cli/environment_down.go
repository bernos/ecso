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

var EnvironmentDownFlags = struct {
	Force cli.BoolFlag
}{
	Force: cli.BoolFlag{
		Name:  "force",
		Usage: "Required. Confirms the environment will be terminated",
	},
}

func NewEnvironmentDownCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return &environmentDownCommandWrapper{ctx, cfg}, nil
	}

	return cli.Command{
		Name:        "down",
		Usage:       "Terminates an ecso environment",
		Description: "Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.",
		ArgsUsage:   "ENVIRONMENT",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			EnvironmentDownFlags.Force,
		},
	}
}

type environmentDownCommandWrapper struct {
	cliCtx *cli.Context
	cfg    *config.Config
}

func (wrapper *environmentDownCommandWrapper) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env   = ctx.Project.Environments[wrapper.cliCtx.Args().First()]
		blue  = ui.NewBannerWriter(w, ui.BlueBold)
		green = ui.NewBannerWriter(w, ui.GreenBold)
	)

	cmd, err := makeEnvironmentCommand(wrapper.cliCtx, ctx.Project, func(env *ecso.Environment) ecso.Command {
		return commands.NewEnvironmentDownCommand(wrapper.cliCtx.Args().First(), wrapper.cfg.EnvironmentAPI(env.Region)).
			WithForce(wrapper.cliCtx.Bool(EnvironmentDownFlags.Force.Name))
	})

	if err != nil {
		return err
	}

	fmt.Fprintf(blue, "Stopping '%s' environment", env.Name)

	if err := cmd.Execute(ctx, r, w); err != nil {
		return err
	}

	fmt.Fprintf(green, "Successfully stopped '%s' environment", env.Name)

	return nil
}

func (wrapper *environmentDownCommandWrapper) Validate(ctx *ecso.CommandContext) error {
	return nil
}
