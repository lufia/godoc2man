package main

import (
	"bytes"
	"io"
	"slices"
)

type ExpWriter struct {
	w      io.Writer
	prefix []byte
}

func NewExpWriter(w io.Writer) *ExpWriter {
	return &ExpWriter{w, nil}
}

func (w *ExpWriter) NeedNextToken(token string) {
	w.prefix = []byte(token)
}

func (w *ExpWriter) Write(data []byte) (n int, err error) {
	n, err = w.writePrefix(data)
	if err != nil {
		return n, err
	}
	m, err := w.w.Write(data)
	if err != nil {
		return n + m, err
	}
	return n + m, nil
}

func (w *ExpWriter) writePrefix(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if data[0] == '\n' {
		w.prefix = nil
		return 0, nil
	}
	if len(w.prefix) == 0 {
		return 0, nil
	}
	i := diffPrefix(data, w.prefix)
	n, err := w.w.Write(w.prefix[:i])
	if err != nil {
		return n, err
	}
	w.prefix = nil
	return n, nil
}

func diffPrefix(data, prefix []byte) int {
	data = bytes.Clone(data)
	slices.Reverse(data)
	prefix = bytes.Clone(prefix)
	slices.Reverse(prefix)
	for n := 0; len(prefix) > 0; n++ {
		if bytes.HasSuffix(data, prefix[:len(prefix)-n]) {
			return n
		}
	}
	return len(prefix)
}

type Heading struct {
	w io.Writer
}

func NewHeading(w io.Writer) *Heading {
	return &Heading{w}
}

func (w *Heading) Write(p []byte) (int, error) {
	var (
		buf bytes.Buffer
		n   int
	)
	for _, c := range p {
		if c == '\n' {
			n++
			continue
		}
		if n > 0 {
			buf.WriteByte(' ')
			n = 0
		}
		buf.WriteByte(c)
	}
	data := bytes.ToUpper(buf.Bytes())
	if n, err := w.w.Write(data); err != nil {
		return n, err
	}
	return len(p), nil
}
