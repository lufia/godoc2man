package main

import (
	"fmt"
	"go/token"
	"strings"
	"testing"
)

func TestPrinterWrite(t *testing.T) {
	var (
		fset token.FileSet
		buf  strings.Builder
	)
	p := NewPrinter(&fset, "example", "1", &buf)
	s := "test"
	fmt.Fprintf(p, "%s", s)
	if v := buf.String(); v != s {
		t.Errorf("Write() = %q; want %q", v, s)
	}
}
