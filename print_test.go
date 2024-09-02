package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrinterWrite(t *testing.T) {
	var buf strings.Builder
	p := NewPrinter("example", 1, &buf)
	s := "test"
	fmt.Fprintf(p, "%s", s)
	if v := buf.String(); v != s {
		t.Errorf("Write() = %q; want %q", v, s)
	}
}
