package ui

import "io"

type prefixWriter struct {
	w io.Writer
	p string
}

func NewWriter(w io.Writer, prefix string) io.Writer {
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
