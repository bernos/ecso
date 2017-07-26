package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	ecsocli "github.com/bernos/ecso/pkg/ecso/cli"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
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

	app := ecsocli.NewApp(cfg, nil, NoopDispatcher())
	app.Setup()

	templates := map[string]string{
		"root":    rootTemplate,
		"command": commandHelpTemplate,
	}

	tmpl := template.Must(parseTemplates(templates))

	if err := tmpl.Execute(os.Stdout, app); err != nil {
		panic(err)
	}
}

// NoopDispatcher creates a dispatcher that does nothing. We need a dispatcher in order to create
// the ecso cli app so that we can introspect all the commands and subcommands when creating the
// documentation
func NoopDispatcher() dispatcher.Dispatcher {
	return dispatcher.DispatcherFunc(func(c dispatcher.CommandFactory, o ...func(*dispatcher.DispatchOptions)) error {
		return nil
	})
}

func parseTemplates(templates map[string]string) (*template.Template, error) {
	var t *template.Template

	for name, body := range templates {
		var tmpl *template.Template

		if t == nil {
			t = template.New(name).Funcs(funcMap)
		}

		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		_, err := tmpl.Parse(body)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

var rootTemplate = `
# ECSO

#### Table of contents
{{range .Commands}}{{$command:=.}}
- [{{.Name}}](#{{.Name}})
{{range .Subcommands}} * [{{.Name}}](#{{$command.Name}}-{{.Name}})
{{end}} {{end}}

{{range .Commands}}
{{template "command" .}}
{{end}}
`

var commandHelpTemplate = `
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
| --{{.Name}} | {{.Usage}} |{{end}}{{end}}{{$Parent:=.}}{{range .Subcommands}} 
<a id="{{$Parent.Name}}-{{.Name}}"></a>
## {{.Name}}

{{.Usage}}{{if .Description}}

{{.Description}}{{end}}

` + "````" + `
ecso {{$Parent.Name}} {{.Name}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
` + "````" + `{{if .VisibleFlags}}

#### Options
| option | usage |
|:---    |:---   |{{range .VisibleFlags}}
| --{{.Name}} | {{.Usage}} |{{end}}{{end}} {{end}}`
