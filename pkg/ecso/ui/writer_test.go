package ui

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	msg := "Hello world"
	buf := &bytes.Buffer{}
	w := NewPrefixWriter(buf, "")

	w.Write([]byte(msg))

	if buf.String() != msg {
		t.Errorf("Want %s, got %s", msg, buf.String())
	}
}

func TestPrefixWriter(t *testing.T) {
	msg := "Hello world"
	buf := &bytes.Buffer{}
	root := NewPrefixWriter(buf, "")
	w := NewPrefixWriter(root, "  ")
	want := "  " + msg

	w.Write([]byte(msg))

	if buf.String() != want {
		t.Errorf("Want '%s', got '%s'", want, buf.String())
	}
}
