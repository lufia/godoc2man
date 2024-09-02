package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestExpWriterWrite(t *testing.T) {
	var buf strings.Builder
	p := NewExpWriter(&buf)
	s := "test"
	fmt.Fprintf(p, "%s", s)
	if v := buf.String(); v != s {
		t.Errorf("Write() = %q; want %q", v, s)
	}
}

func TestPrinterWriteWithPrefix(t *testing.T) {
	tests := map[string]struct {
		prefix string
		format string
		want   string
	}{
		"empty": {" ", "", ""},
		"equal": {" - ", " - godoc", " - godoc"},
		"less1": {" - ", "- godoc", " - godoc"},
		"less2": {" - ", " = godoc", " - = godoc"},
		"none":  {" ", "godoc", " godoc"},
		"reset": {" ", "\n.SH", "\n.SH"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf strings.Builder
			w := NewExpWriter(&buf)
			w.NeedNextToken(tt.prefix)
			fmt.Fprintf(w, tt.format)
			if v := buf.String(); v != tt.want {
				t.Errorf("Write() = %q; want %q", v, tt.want)
			}
		})
	}
}
