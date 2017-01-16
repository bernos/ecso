package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type WriterFunc func(p []byte) (n int, err error)

func (w WriterFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

type Logger interface {
	Error(format string, a ...interface{})
	Info(format string, a ...interface{})
	ErrWriter() io.Writer
}

func New(w io.Writer) Logger {
	return &log{w}
}

type log struct {
	w io.Writer
}

func (l *log) Error(format string, a ...interface{}) {
	l.writeError(fmt.Sprintf(format, a))
}

func (l *log) Info(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(l.w, bold("Info: ")+format, a...)
}
func (l *log) writeError(msg string) (n int, err error) {
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.FgRed, color.Bold).SprintFunc()

	return fmt.Fprintf(l.w, "%s %s", bold("Error:"), red(msg))
}

func (l *log) ErrWriter() io.Writer {
	return WriterFunc(func(p []byte) (n int, err error) {
		return l.writeError(string(p))
	})
}
