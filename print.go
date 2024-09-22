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
	p.writeBugs(pkg.Notes["BUG"])
}

var optionDef = strings.TrimSpace(`
.de OPT
.TP
\fB\-\\$1\fR=\fI\\$2\fR
.shift 2
\\$*
..
`)

func (p *Printer) writeHeader(pkg *doc.Package, flags []*Flag) {
	name := path.Base(p.pkgPath)
	fmt.Fprintf(p, ".TH %s %d\n", name, p.section)
	fmt.Fprintln(p, ".SH NAME")
	s := pkg.Synopsis(pkg.Doc)
	s = strings.TrimPrefix(s, name)
	s = strings.TrimSpace(s)
	fmt.Fprintf(p, "%s \\- %s\n", name, s)
	if len(flags) > 0 {
		fmt.Fprintln(p, ".SH OPTIONS")
		fmt.Fprintln(p, optionDef)
		for _, flg := range flags {
			fmt.Fprintln(p, ".OPT", flg.Name, strings.ToUpper(flg.Placeholder), flg.Usage)
		}
	}
	fmt.Fprintln(p, ".SH OVERVIEW")
}

func (p *Printer) writeContent(content []comment.Block, depth int) {
	for _, c := range content {
		switch c := c.(type) {
		case *comment.Heading:
			w := NewHeading(p)
			fmt.Fprintf(w, ".SH %s", Text(c.Text))
			fmt.Fprintln(p, "")
		case *comment.Paragraph:
			if depth == 0 {
				fmt.Fprintln(p, ".PP")
			}
			fmt.Fprintf(p, "%+s", Text(c.Text))
		case *comment.Code:
			fmt.Fprintln(p, ".PP")
			fmt.Fprintln(p, ".EX")
			fmt.Fprintln(p, ".in +4n")
			fmt.Fprintf(p, "%s\n", roff.Str(c.Text))
			fmt.Fprintln(p, ".in")
			fmt.Fprintln(p, ".EE")
		case *comment.List:
			for _, item := range c.Items {
				symbol := roff.Bullet
				if item.Number != "" {
					symbol = item.Number + "."
				}
				fmt.Fprintf(p, ".IP %s 4\n", symbol)
				p.writeContent(item.Content, depth+1)
			}
		}
	}
}

func (p *Printer) writeBugs(a []*doc.Note) {
	if len(a) == 0 {
		return
	}
	fmt.Fprintln(p, ".SH BUGS")
	for _, n := range a {
		fmt.Fprintln(p, ".PP")
		fmt.Fprintln(p, n.Body)
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
			fmt.Fprintf(w, "%s\n", roff.Str(v))
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
