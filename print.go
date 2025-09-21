package main

import (
	"fmt"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
	"io"
	"iter"
	"path"
	"strings"

	"github.com/lufia/godoc2man/internal/roff"
)

type Printer struct {
	fset    *token.FileSet
	pkgPath string
	section string
	w       io.Writer
	err     error
}

func NewPrinter(fset *token.FileSet, pkgPath, section string, w io.Writer) *Printer {
	return &Printer{fset, pkgPath, section, w, nil}
}

func (p *Printer) Err() error {
	return p.err
}

func (p *Printer) Command(pkg *doc.Package, d *comment.Doc, flags []*Flag) {
	p.writeHeader(pkg, flags)
	p.writeContent(d.Content, 0, false)
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
	fmt.Fprintf(p, ".TH %s %s\n", name, p.section)
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

func (p *Printer) writeContent(content []comment.Block, depth int, cont bool) {
	for _, c := range content {
		switch c := c.(type) {
		case *comment.Heading:
			w := NewHeading(p)
			fmt.Fprintf(w, ".SH %s", Text(c.Text))
			fmt.Fprintln(p, "")
		case *comment.Paragraph:
			if depth == 0 && !cont {
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
				p.writeContent(item.Content, depth+1, false)
			}
		}
		cont = false
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

func (p *Printer) Library(pkg *doc.Package, d *comment.Doc) {
	name := path.Base(p.pkgPath)
	fmt.Fprintf(p, ".TH %s %s\n", name, p.section)
	fmt.Fprintln(p, ".SH NAME")
	s := pkg.Synopsis(pkg.Doc)
	s = strings.TrimPrefix(s, name)
	s = strings.TrimSpace(s)
	fmt.Fprintf(p, "%s \\- %s\n", name, s)

	fmt.Fprintln(p, ".SH SYNOPSIS")
	fmt.Fprintln(p, ".nf")
	fmt.Fprintf(p, ".B \"import \\(dq%s\\(dq\"\n", p.pkgPath)
	fmt.Fprintln(p, ".sp")
	for _, v := range pkg.Vars {
		writeVar(p, p.fset, v)
	}
	ndef := len(pkg.Vars)
	if ndef > 0 && len(pkg.Types) > 0 {
		fmt.Fprint(p, "\n")
		ndef = 0
	}
	ndef += len(pkg.Types)
	for _, t := range pkg.Types {
		writeType(p, p.fset, t)
	}
	if ndef > 0 && len(pkg.Funcs) > 0 {
		fmt.Fprint(p, "\n")
		ndef = 0
	}
	ndef += len(pkg.Funcs)
	for _, f := range pkg.Funcs {
		fmt.Fprint(p, `.BI "`)
		writeFunc(p, p.fset, f)
		fmt.Fprintln(p, `\"`)
	}
	fmt.Fprintln(p, ".fi")
	fmt.Fprintln(p, ".SH DESCRIPTION")
	p.writeContent(d.Content, 0, false)
	if len(pkg.Vars) > 0 {
		fmt.Fprintln(p, ".PP")
	}
	fmt.Fprintln(p, ".SS Variables")
	for _, t := range pkg.Vars {
		p.writeSymbolDoc(t.Doc, t.Names[0])
	}
	if len(pkg.Types) > 0 {
		fmt.Fprintln(p, ".PP")
	}
	fmt.Fprintln(p, ".SS Types")
	for _, t := range pkg.Types {
		p.writeSymbolDoc(t.Doc, t.Name)
	}
	if len(pkg.Funcs) > 0 {
		fmt.Fprintln(p, ".PP")
	}
	fmt.Fprintln(p, ".SS Functions")
	for _, t := range pkg.Funcs {
		s := t.Doc
		if strings.HasPrefix(s, t.Name) {
			fmt.Fprintln(p, ".BR", t.Name, "()")
			s = s[len(t.Name):]
		}
		var parser comment.Parser
		doc := parser.Parse(strings.TrimSpace(s))
		p.writeContent(doc.Content, 0, true)
		fmt.Fprintln(p, ".PP")
	}
	p.writeBugs(pkg.Notes["BUG"])
}

func writeType(w io.Writer, fset *token.FileSet, t *doc.Type) error {
	if err := format.Node(w, fset, t.Decl); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "\n")
	for f := range mergeSlice(t.Funcs, t.Methods) {
		fmt.Fprint(w, `.BI "`)
		writeFunc(w, fset, f)
		fmt.Fprintln(w, `\"`)
	}
	_, err = fmt.Fprint(w, "\n")
	return err
}

func mergeSlice[S ~[]E, E any](s ...S) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, a := range s {
			for _, v := range a {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func writeVar(w io.Writer, fset *token.FileSet, v *doc.Value) error {
	if err := format.Node(w, fset, v.Decl); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}

func writeFunc(w io.Writer, fset *token.FileSet, f *doc.Func) error {
	x := *f.Decl
	x.Body = nil
	return format.Node(w, fset, &x)
}

func (p *Printer) writeSymbolDoc(s, name string) {
	before, rest, ok := hasPrefix(s, name)
	if ok {
		fmt.Fprintln(p, before)
		fmt.Fprintln(p, ".BR", name)
		s = rest
	}
	var parser comment.Parser
	doc := parser.Parse(strings.TrimSpace(s))
	p.writeContent(doc.Content, 0, true)
	fmt.Fprint(p, "\n")
}

func hasPrefix(s, name string) (before, rest string, ok bool) {
	switch {
	case strings.HasPrefix(s, "The "):
		before = "The"
		s = s[4:]
	case strings.HasPrefix(s, "An "):
		before = "An"
		s = s[3:]
	case strings.HasPrefix(s, "A "):
		before = "A"
		s = s[2:]
	}
	if strings.HasPrefix(s, name+" ") {
		return before, s[len(name)+1:], true
	}
	return "", "", false
}
