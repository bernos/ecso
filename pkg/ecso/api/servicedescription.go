package api

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso/ui"
)

type ServiceDescription struct {
	Name                     string
	URL                      string
	CloudFormationConsoleURL string
	CloudWatchLogsConsoleURL string
	ECSConsoleURL            string
	CloudFormationOutputs    map[string]string
}

func (s *ServiceDescription) WriteTo(w io.Writer) (int64, error) {
	buf := &bytes.Buffer{}
	blue := ui.NewBannerWriter(buf, ui.BlueBold)
	pw := ui.NewPrefixWriter(buf, "  ")
	dt := ui.NewDefinitionWriter(pw, ":")

	fmt.Fprintf(blue, "Details of the '%s' service:", s.Name)
	fmt.Fprintf(dt, "CloudFormation console:%s", s.CloudFormationConsoleURL)
	fmt.Fprintf(dt, "CloudWatch logs:%s", s.CloudWatchLogsConsoleURL)
	fmt.Fprintf(dt, "ECS console:%s", s.ECSConsoleURL)

	if s.URL != "" {
		fmt.Fprintf(dt, "Service URL:%s", s.URL)
	}

	fmt.Fprintf(blue, "CloudFormation Outputs:")

	for k, v := range s.CloudFormationOutputs {
		fmt.Fprintf(dt, "%s:%s", k, v)
	}

	n, err := w.Write(buf.Bytes())

	return int64(n), err
}
