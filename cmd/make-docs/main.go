package main

import (
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

var CommandHelpTemplate = `
## {{.Name}}

{{.Usage}}{{if .Description}}

{{.Description}}{{end}}

` + "````" + `
{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `{{if .Category}}

#### Category
{{.Category}}{{end}}{{if .VisibleFlags}}

#### Options
{{range .VisibleFlags}}- {{.}}
{{end}}{{end}}`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `
# {{.Name}}

{{.Usage}}

` + "````" + `
{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `

#### Commands{{range .VisibleCategories}}{{if .Name}}
{{.Name}}:{{end}}{{range .VisibleCommands}}
- {{.Name}}{{with .ShortName}}, {{.}}{{end}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}

#### Options
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}`

func main() {
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate

	dispatcher := ecso.DispatcherFunc(func(c ecso.Command, o ...func(*ecso.DispatchOptions)) error {
		return nil
	})

	app := cmd.NewApp("", dispatcher)

	for _, command := range app.Commands {
		app.Run([]string{"ecso", command.Name, "--help"})

		for _, sub := range command.Subcommands {
			app.Run([]string{"ecso", command.Name, sub.Name, "--help"})
		}
	}
}

func printHelp(out io.Writer, templ string, data interface{}) {
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	w := tabwriter.NewWriter(out, 0, 8, 1, '\t', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))
	err := t.Execute(w, data)
	if err != nil {
		// If the writer is closed, t.Execute will fail, and there's nothing
		// we can do to recover.
		if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
			// fmt.Fprintf(ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}
	w.Flush()
}
