package api

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso/ui"
)

type EnvironmentDescription struct {
	Name                     string
	CloudFormationConsoleURL string
	CloudWatchLogsConsoleURL string
	ECSConsoleURL            string
	ECSClusterBaseURL        string
	CloudFormationOutputs    map[string]string
}

func (env *EnvironmentDescription) WriteTo(w io.Writer) (int64, error) {
	buf := &bytes.Buffer{}
	blue := ui.NewBannerWriter(buf, ui.BlueBold)
	pw := ui.NewPrefixWriter(buf, "  ")
	dt := ui.NewDefinitionWriter(pw, ":")

	fmt.Fprintf(blue, "Details of the '%s' environment:", env.Name)
	fmt.Fprintf(dt, "CloudFormation console:%s", env.CloudFormationConsoleURL)
	fmt.Fprintf(dt, "CloudWatch logs:%s", env.CloudWatchLogsConsoleURL)
	fmt.Fprintf(dt, "ECS console:%s", env.ECSConsoleURL)
	fmt.Fprintf(dt, "ECS base URL:%s", env.ECSClusterBaseURL)

	fmt.Fprintf(blue, "CloudFormation Outputs:")

	for k, v := range env.CloudFormationOutputs {
		fmt.Fprintf(dt, "%s:%s", k, v)
	}

	n, err := w.Write(buf.Bytes())

	return int64(n), err
}
