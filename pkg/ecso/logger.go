package ecso

import (
	"fmt"
	"io"

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

func (l *log) Infof(format string, a ...interface{}) {
	fmt.Fprintf(l.w, bold("Info: ")+format+"\n", a...)
}

func (l *log) writeError(msg string) (n int, err error) {
	return fmt.Fprintf(l.w, "%s %s", redBold("Error:"), red("%s", msg))
}

func (l *log) ErrWriter() io.Writer {
	return writerFunc(func(p []byte) (n int, err error) {
		return l.writeError(string(p))
	})
}
