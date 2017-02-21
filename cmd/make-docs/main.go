package main

import (
	"fmt"
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
<a id="{{.Name}}"></a>
## {{.HelpName}}

{{.Usage}}{{if .Description}}

{{.Description}}{{end}}

` + "````" + `
{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `{{if .Category}}

#### Category
{{.Category}}{{end}}{{if .VisibleFlags}}

#### Options
| option | usage |
|:---    |:---   |{{range .VisibleFlags}}
| --{{.Name}} | {{.Usage}} |{{end}}{{end}}
`

// SubcommandHelpTemplate is the text template for the subcommand help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var SubcommandHelpTemplate = `
<a id="{{.Name}}"></a>
# {{.HelpName}}

{{.Usage}}

` + "````" + `
{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `

#### Commands{{range .VisibleCategories}}{{if .Name}}
{{.Name}}:{{end}}
| Name  | Description |
|:---   |:---         |{{range .VisibleCommands}}
| --{{.Name}}{{with .ShortName}}, --{{.}}{{end}} | {{.Usage}} | {{end}}
{{end}}{{if .VisibleFlags}}

#### Options
| option | usage |
|:---    |:---   |{{range .VisibleFlags}}
| --{{.Name}} | {{.Usage}} |{{end}}{{end}}
`

func main() {
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate

	dispatcher := ecso.DispatcherFunc(func(c ecso.Command, o ...func(*ecso.DispatchOptions)) error {
		return nil
	})

	commands := make([][]string, 0)

	app := cmd.NewApp("", dispatcher)

	fmt.Println("# ECSO ")
	fmt.Printf("\n#### Table of contents\n\n")

	for _, command := range app.Commands {
		fmt.Printf("- [%s](#%s)\n", command.Name, command.Name)
		commands = append(commands, []string{"ecso", command.Name, "--help"})

		for _, sub := range command.Subcommands {
			fmt.Printf("  * [%s](#%s)\n", sub.Name, sub.Name)
			commands = append(commands, []string{"ecso", command.Name, sub.Name, "--help"})
		}
	}

	for _, c := range commands {
		app = cmd.NewApp("", dispatcher)
		app.Run(c)
		// time.Sleep(time.Second)
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
