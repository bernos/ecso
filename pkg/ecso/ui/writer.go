package ui

import "io"

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
	return bw.output.Write([]byte(bw.sprintf("\n%s\n\n", string(p))))
}
