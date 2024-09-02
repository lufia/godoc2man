package roff

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	Str("untyped string")
	type Tstr string
	Str(Tstr("string underlying"))
}

func TestStrString(t *testing.T) {
	s := Str("test")
	w := "test"
	if v := s.String(); v != w {
		t.Errorf("String() = %v; want %v", v, w)
	}
}

func TestStrFormat_s(t *testing.T) {
	testFormat(t, "%s", map[string]string{
		"text":     "text",
		" text":    "text",
		"text ":    "text",
		"an apple": "an apple",
		`\`:        `\(rs`,
		"-":        `\-`,
		`"`:        `\(dq`,
	})
	testFormat(t, "% s", map[string]string{
		" text":    " text",
		"text ":    "text ",
		"an apple": "an apple",
	})
}

func TestStrFormat_S(t *testing.T) {
	testFormat(t, "%S", map[string]string{
		"text":     "text",
		" text":    "text",
		"text ":    "text",
		"an apple": "an apple",
		`\`:        `\`,
		"-":        "-",
		`"`:        `"`,
	})
	testFormat(t, "% S", map[string]string{
		" text":    " text",
		"text ":    "text ",
		"an apple": "an apple",
	})
}

func TestStrFormat_quote(t *testing.T) {
	testFormat(t, "%q", map[string]string{
		"text":     `"text"`,
		" text":    `"text"`,
		"text ":    `"text"`,
		"an apple": `"an apple"`,
		`\`:        `"\(rs"`,
		"-":        `"\-"`,
		`"`:        `"\(dq"`,
	})
	testFormat(t, "% q", map[string]string{
		" text":    `" text"`,
		"text ":    `"text "`,
		"an apple": `"an apple"`,
	})
}

func testFormat(t *testing.T, format string, tests map[string]string) {
	t.Helper()
	for s, w := range tests {
		if v := fmt.Sprintf(format, Str(s)); v != w {
			t.Errorf("Format(%q, %q) = %q; want %q", format, s, v, w)
		}
	}
}
