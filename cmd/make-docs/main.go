package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	ecsocli "github.com/bernos/ecso/pkg/ecso/cli"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

var (
	funcMap = template.FuncMap{
		"join": strings.Join,
	}
)

func main() {
	cfg, err := config.NewConfig("")

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	app := ecsocli.NewApp(cfg, NoopDispatcher())

	WriteTo(app, os.Stdout)
}

// NoopDispatcher creates a dispatcher that does nothing. We need a dispatcher in order to create
// the ecso cli app so that we can introspect all the commands and subcommands when creating the
// documentation
func NoopDispatcher() dispatcher.Dispatcher {
	return dispatcher.DispatcherFunc(func(c dispatcher.CommandFactory, o ...func(*dispatcher.DispatchOptions)) error {
		return nil
	})
}

func WriteTo(app *cli.App, w io.Writer) {
	app.Setup()

	fmt.Fprintln(w, "# ECSO ")
	fmt.Fprintf(w, "\n#### Table of contents\n\n")

	for _, command := range app.Commands {
		fmt.Fprintf(w, "- [%s](#%s)\n", command.Name, command.Name)

		for _, sub := range command.Subcommands {
			fmt.Fprintf(w, "  * [%s](#%s-%s)\n", sub.Name, command.Name, sub.Name)
		}
	}

	for _, command := range app.Commands {
		WriteCommand(&Command{&command, nil}, w)

		for _, sub := range command.Subcommands {
			WriteCommand(&Command{&sub, &command}, w)
		}
	}
}

func WriteCommand(c *Command, w io.Writer) {
	var t *template.Template

	if c.Parent != nil {
		t = SubCommandHelpTemplate
	} else {
		t = CommandHelpTemplate
	}

	if err := t.Execute(w, c); err != nil {
		panic(err)
	}
}

type Command struct {
	*cli.Command
	Parent *cli.Command
}

var SubCommandHelpTemplate = template.Must(template.New("SubCommandHelp").Funcs(funcMap).Parse(`
<a id="{{.Parent.Name}}-{{.Name}}"></a>
## {{.Name}}

{{.Usage}}{{if .Description}}

{{.Description}}{{end}}

` + "````" + `
ecso {{.Parent.Name}} {{.Name}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `{{if .VisibleFlags}}

#### Options
| option | usage |
|:---    |:---   |{{range .VisibleFlags}}
| --{{.Name}} | {{.Usage}} |{{end}}{{end}}
`))

var CommandHelpTemplate = template.Must(template.New("CommandHelp").Funcs(funcMap).Parse(`
<a id="{{.Name}}"></a>
# {{.Name}}

{{.Usage}}

` + "````" + `
ecso {{.Name}}{{if .Subcommands}} <command>{{end}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `{{if .Subcommands}}

#### Commands
| Name  | Description |
|:---   |:---         |{{$p := .Name}}{{range .Subcommands}}
| [{{.Name}}](#{{$p}}-{{.Name}}) | {{.Usage}} | {{end}}
{{end}}{{if .VisibleFlags}}

#### Options
| option | usage |
|:---    |:---   |{{range .VisibleFlags}}
| --{{.Name}} | {{.Usage}} |{{end}}{{end}}
`))
