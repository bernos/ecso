package ui

import (
	"fmt"
	"io"
	"strings"
)

type writerFunc func([]byte) (int, error)

func (fn writerFunc) Write(p []byte) (int, error) {
	return fn(p)
}

type prefixWriter struct {
	w io.Writer
	p string
}

func NewPrefixWriter(w io.Writer, prefix string) io.Writer {
	switch parent := w.(type) {
	case *prefixWriter:
		return &prefixWriter{w: w, p: parent.p + prefix}
	default:
		return &prefixWriter{w: w, p: prefix}
	}
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	return pw.w.Write(append([]byte(pw.p), p...))
}

type bannerWriter struct {
	output  io.Writer
	sprintf func(string, ...interface{}) string
}

func NewBannerWriter(w io.Writer, color Color) io.Writer {
	return &bannerWriter{
		output:  w,
		sprintf: colors[color],
	}
}

func (bw *bannerWriter) Write(p []byte) (int, error) {
	return bw.output.Write([]byte(bw.sprintf("\n%s\n\n", p)))
}

func NewInfoWriter(w io.Writer) io.Writer {
	return writerFunc(func(p []byte) (int, error) {
		return w.Write([]byte(fmt.Sprintf("%s %s\n", bold("Info:"), p)))
	})
}

func NewErrWriter(w io.Writer) io.Writer {
	return writerFunc(func(p []byte) (int, error) {
		return w.Write([]byte(fmt.Sprintf("%s %s\n", redBold("Error:"), red("%s", p))))
	})
}

type definitionWriter struct {
	output    io.Writer
	delimiter string
}

func NewDefinitionWriter(w io.Writer, delimiter string) io.Writer {
	return &definitionWriter{
		output:    w,
		delimiter: delimiter,
	}
}

func (w *definitionWriter) Write(p []byte) (int, error) {
	tokens := strings.Split(string(p), w.delimiter)
	label := tokens[0] + ":"
	value := ""

	if len(tokens) > 1 {
		value = strings.Join(tokens[1:], w.delimiter)
	}

	if n, err := w.output.Write([]byte(fmt.Sprintf("%s\n", bold(label)))); err != nil {
		return n, err
	}

	if n, err := w.output.Write([]byte(fmt.Sprintf("  %s\n", value))); err != nil {
		return n, err
	}

	return len(p), nil
}

type TableWriter struct {
	output    io.Writer
	headers   []string
	rows      [][]string
	delimeter string
}

func NewTableWriter(w io.Writer, delimeter string) *TableWriter {
	return &TableWriter{
		output:    w,
		headers:   make([]string, 0),
		rows:      make([][]string, 0),
		delimeter: delimeter,
	}
}

func (t *TableWriter) WriteHeader(p []byte) (int, error) {
	if len(t.headers) != 0 {
		return 0, fmt.Errorf("Multiple calls to TableWriter.WriteHeader")
	}

	t.headers = strings.Split(string(p), t.delimeter)

	return len(p), nil
}

func (t *TableWriter) Write(p []byte) (int, error) {
	t.rows = append(t.rows, strings.Split(string(p), t.delimeter))
	return len(p), nil
}

func (t *TableWriter) Flush() (int, error) {
	format := ""
	n := 0

	for i, h := range t.headers {
		l := len(h)

		for _, r := range t.rows {
			if len(r[i]) > l {
				l = len(r[i])
			}
		}

		format = format + fmt.Sprintf("%%-%ds  ", l)
	}

	format = format + "\n"

	x, err := fmt.Fprintf(t.output, format, toInterfaceSlice(t.headers)...)
	if err != nil {
		return n + x, err
	}

	n = n + x

	for i := range t.rows {
		x, err := fmt.Fprintf(t.output, format, toInterfaceSlice(t.rows[i])...)
		if err != nil {
			return n + x, err
		}

		n = n + x
	}

	return n, nil
}

func toInterfaceSlice(xs []string) []interface{} {
	ys := make([]interface{}, len(xs))

	for i := range xs {
		ys[i] = xs[i]
	}

	return ys
}
