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

	str := fmt.Sprintf("%s\n  %s\n", bold(label), value)

	return w.output.Write([]byte(str))
}
