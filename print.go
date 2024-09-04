package main

import (
	"fmt"
	"go/doc"
	"go/doc/comment"
	"io"
	"path"
	"strings"

	"github.com/lufia/godoc2man/internal/roff"
)

type Printer struct {
	pkgPath string
	section int
	w       io.Writer
	err     error
}

func NewPrinter(pkgPath string, section int, w io.Writer) *Printer {
	return &Printer{pkgPath, section, w, nil}
}

func (p *Printer) Err() error {
	return p.err
}

func (p *Printer) Command(pkg *doc.Package, d *comment.Doc, flags []*Flag) {
	p.writeHeader(pkg, flags)
	p.writeContent(d.Content, 0)
}

func (p *Printer) writeHeader(pkg *doc.Package, flags []*Flag) {
	fmt.Fprintf(p, ".TH %s %d\n", p.pkgPath, p.section)
	fmt.Fprintf(p, ".SH NAME\n")
	name := path.Base(p.pkgPath)
	s := pkg.Synopsis(pkg.Doc)
	s = strings.TrimPrefix(s, name)
	s = strings.TrimSpace(s)
	fmt.Fprintf(p, "%s \\- %s\n", name, s)
	if len(flags) > 0 {
		fmt.Fprintln(p, ".SH OPTIONS")
		for _, flg := range flags {
			fmt.Fprintln(p, ".TP")
			fmt.Fprintf(p, ".BI \"\\-%s \" %s\n", flg.Name, flg.Placeholder)
			fmt.Fprintln(p, flg.Usage)
		}
	}
	fmt.Fprintf(p, ".SH OVERVIEW\n")
}

const bullet = `\(bu`

func (p *Printer) writeContent(content []comment.Block, depth int) {
	for _, c := range content {
		switch c := c.(type) {
		case *comment.Heading:
			w := NewHeading(p)
			fmt.Fprintf(w, ".SH %s", Text(c.Text))
			fmt.Fprintln(p, "")
		case *comment.Paragraph:
			if depth == 0 {
				fmt.Fprintf(p, ".PP\n")
			}
			fmt.Fprintf(p, "%+s", Text(c.Text))
		case *comment.Code:
			fmt.Fprintf(p, ".PP\n")
			fmt.Fprintf(p, ".EX\n")
			fmt.Fprintf(p, ".in +4n\n")
			fmt.Fprintf(p, "%s\n", roff.Str(c.Text))
			fmt.Fprintf(p, ".in\n")
			fmt.Fprintf(p, ".EE\n")
		case *comment.List:
			for _, item := range c.Items {
				symbol := bullet
				if item.Number != "" {
					symbol = item.Number + "."
				}
				fmt.Fprintf(p, ".IP %s 4\n", symbol)
				p.writeContent(item.Content, depth+1)
			}
		}
	}
}

func (p *Printer) Write(data []byte) (n int, err error) {
	if p.err != nil {
		return 0, p.err
	}
	return p.w.Write(data)
}

type Text []comment.Text

func (t Text) Format(f fmt.State, c rune) {
	w := NewExpWriter(f)
	format := "%"
	if f.Flag('+') {
		format += "+"
	}
	format += string(c)

	trailing := false
	for _, v := range t {
		switch v := v.(type) {
		case comment.Plain:
			if trailing {
				if strings.HasPrefix(string(v), " ") {
					fmt.Fprint(w, "\n")
				} else {
					fmt.Fprint(w, " ")
				}
				trailing = false
			}
			s := fmt.Sprintf("%s", roff.Str(v))
			fmt.Fprintf(w, "%s\n", BreakString(s))
		case comment.Italic:
			if f.Flag('+') {
				fmt.Fprintf(w, ".I ")
			}
			fmt.Fprintf(w, "%s\n", roff.Str(v))
		case *comment.Link:
			if f.Flag('+') {
				fmt.Fprintf(w, ".UR %q\n", roff.Str(v.URL))
			}
			fmt.Fprintf(w, format, Text(v.Text))
			if f.Flag('+') {
				fmt.Fprintf(w, ".UE")
				trailing = true
			}
		case *comment.DocLink:
			if f.Flag('+') {
				u := v.DefaultURL("https://pkg.go.dev")
				fmt.Fprintf(w, "\n.UR %q\n", roff.Str(u))
			}
			fmt.Fprintf(w, format, Text(v.Text))
			if f.Flag('+') {
				fmt.Fprintf(w, ".UE")
				trailing = true
			}
		}
	}
}
