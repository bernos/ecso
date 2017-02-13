package ecso

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	red     = color.New(color.FgRed).SprintfFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintfFunc()

	bold = color.New(color.Bold).SprintfFunc()
)

type writerFunc func(p []byte) (n int, err error)

func (w writerFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

type Logger interface {
	Child() Logger
	Errorf(format string, a ...interface{})
	Printf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	ErrWriter() io.Writer
	Writer() io.Writer
}

func NewLogger(w io.Writer) Logger {
	return &log{w, ""}
}

type log struct {
	w      io.Writer
	prefix string
}

func (l *log) Child() Logger {
	return &log{l.w, l.prefix + "  "}
}

func (l *log) Errorf(format string, a ...interface{}) {
	l.writeError(fmt.Sprintf(format, a))
}

func (l *log) Infof(format string, a ...interface{}) {
	fmt.Fprintf(l.w, l.prefix+bold("Info: ")+format+"\n", a...)
}

func (l *log) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.w, l.prefix+format, a...)
}

func (l *log) writeError(msg string) (n int, err error) {
	return fmt.Fprintf(l.w, "%s%s %s", l.prefix, redBold("Error:"), red("%s", msg))
}

func (l *log) ErrWriter() io.Writer {
	return writerFunc(func(p []byte) (n int, err error) {
		return l.writeError(string(p))
	})
}

func (l *log) Writer() io.Writer {
	return l.w
}
