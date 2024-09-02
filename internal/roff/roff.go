// Package roff provides utilities for roff text.
package roff

import (
	"fmt"
	"strings"
)

var escaper = strings.NewReplacer(
	`\`, `\(rs`,
	"-", `\-`,
	`"`, `\(dq`,
)

// String represents a string
type String struct {
	s string
}

var (
	_ fmt.Stringer  = (*String)(nil)
	_ fmt.Formatter = (*String)(nil)
)

func Str[T ~string](s T) String {
	return String{string(s)}
}

func (s String) String() string {
	return s.s
}

func (s String) Format(f fmt.State, c rune) {
	v := s.s
	if !f.Flag(' ') {
		v = strings.TrimSpace(v)
	}
	switch c {
	case 's':
		fmt.Fprintf(f, "%s", escaper.Replace(v))
	case 'S':
		fmt.Fprintf(f, "%s", v)
	case 'q':
		fmt.Fprintf(f, `"%s"`, escaper.Replace(v))
	}
}
