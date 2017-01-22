package ecso

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	blue     = color.New(color.FgBlue).SprintfFunc()
	blueBold = color.New(color.FgBlue, color.Bold).SprintfFunc()

	green     = color.New(color.FgGreen).SprintfFunc()
	greenBold = color.New(color.FgGreen, color.Bold).SprintfFunc()

	red     = color.New(color.FgRed).SprintfFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintfFunc()

	bold = color.New(color.Bold).SprintfFunc()
)

type writerFunc func(p []byte) (n int, err error)

func (w writerFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

type Logger interface {
	BannerBlue(format string, a ...interface{})
	BannerGreen(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Fatalf(format string, a ...interface{})
	Printf(format string, a ...interface{})
	PrefixPrintf(prefix string) func(string, ...interface{})
	Infof(format string, a ...interface{})
	ErrWriter() io.Writer
}

func NewLogger(w io.Writer) Logger {
	return &log{w}
}

type log struct {
	w io.Writer
}

func (l *log) BannerBlue(format string, a ...interface{}) {
	fmt.Fprintf(l.w, "\n%s\n\n", blueBold(format, a...))
}

func (l *log) BannerGreen(format string, a ...interface{}) {
	fmt.Fprintf(l.w, "\n%s\n\n", greenBold(format, a...))
}

func (l *log) Errorf(format string, a ...interface{}) {
	l.writeError(fmt.Sprintf(format, a))
}

func (l *log) Fatalf(format string, a ...interface{}) {
	l.Errorf(format, a...)
	os.Exit(1)
}

func (l *log) Infof(format string, a ...interface{}) {
	fmt.Fprintf(l.w, bold("Info: ")+format+"\n", a...)
}

func (l *log) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.w, format, a...)
}

func (l *log) PrefixPrintf(prefix string) func(string, ...interface{}) {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(l.w, prefix+format, a...)
	}
}

func (l *log) writeError(msg string) (n int, err error) {
	return fmt.Fprintf(l.w, "%s %s", redBold("Error:"), red("%s", msg))
}

func (l *log) ErrWriter() io.Writer {
	return writerFunc(func(p []byte) (n int, err error) {
		return l.writeError(string(p))
	})
}
