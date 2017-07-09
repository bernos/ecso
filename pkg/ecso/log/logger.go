package log

import (
	"fmt"
	"io"
	"sync"

	"github.com/fatih/color"
)

var (
	bold    = color.New(color.Bold).SprintfFunc()
	red     = color.New(color.FgRed).SprintfFunc()
	redBold = color.New(color.FgRed, color.Bold).SprintfFunc()
)

type writerFunc func(p []byte) (n int, err error)

func (w writerFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

// Logger writes log messages to a writer. Loggers enable heirarchical logging
// via the Child() method
type Logger interface {
	Child() Logger
	Errorf(format string, a ...interface{})
	Printf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	// ErrWriter() io.Writer
	// Writer() io.Writer
}

// NewLogger creates a new logger that will write to the provided writer. all
// messages will be prefixed with the provided prefix string
func NewLogger(w io.Writer, prefix string) Logger {
	return &log{
		w:      w,
		prefix: prefix,
	}
}

type log struct {
	mu     sync.Mutex // ensures atomic writes
	w      io.Writer
	prefix string
}

func (l *log) Child() Logger {
	return &log{
		w:      l.w,
		prefix: l.prefix + "  ",
	}
}

func (l *log) Errorf(format string, a ...interface{}) {
	l.writeError(fmt.Sprintf(format, a))
}

func (l *log) Infof(format string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.w, l.prefix+bold("Info: ")+format+"\n", a...)
}

func (l *log) Printf(format string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.w, l.prefix+format, a...)
}

func (l *log) writeError(msg string) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return fmt.Fprintf(l.w, "%s%s %s", l.prefix, redBold("Error:"), red("%s", msg))
}

func (l *log) ErrWriter() io.Writer {
	return writerFunc(func(p []byte) (n int, err error) {
		return l.writeError(string(p))
	})
}

func (l *log) Writer() io.Writer {
	return writerFunc(func(p []byte) (n int, err error) {
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.w.Write(p)
	})
}
