package api

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso/ui"
)

type ServiceVersion struct {
	Service string
	Label   string
}

type ServiceVersionList []*ServiceVersion

func (l ServiceVersionList) WriteTo(w io.Writer) (int64, error) {
	tw := ui.NewTableWriter(w, "|")
	tw.WriteHeader([]byte("SERVICE|VERSION"))

	for _, v := range l {
		row := fmt.Sprintf("%s|%s", v.Service, v.Label)
		tw.Write([]byte(row))
	}

	n, err := tw.Flush()

	return int64(n), err
}
