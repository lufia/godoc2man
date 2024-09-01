package main

import (
	"fmt"
	"go/doc"
	"go/doc/comment"
	"io"
	"path"
	"strings"
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

func (p *Printer) Command(pkg *doc.Package, d *comment.Doc) {
	p.writeHeader(pkg)
	p.writeContent(d.Content, 0)
}

func (p *Printer) writeHeader(pkg *doc.Package) {
	p.writeString(".TH %s %d\n", p.pkgPath, p.section)
	p.writeString(".SH NAME\n")
	name := path.Base(p.pkgPath)
	s := pkg.Synopsis(pkg.Doc)
	s = strings.TrimPrefix(s, name)
	s = strings.TrimSpace(s)
	p.writeString("%s \\- %s\n", name, s)
	p.writeString(".SH OVERVIEW\n")
}

const bullet = `\(bu`

func (p *Printer) writeContent(content []comment.Block, depth int) {
	for _, c := range content {
		switch c := c.(type) {
		case *comment.Heading:
			p.writeString(".SH %s\n", HeadingText(c.Text))
		case *comment.Paragraph:
			if depth == 0 {
				p.writeString(".PP\n")
			}
			p.writeString("%s", ParagraphText(c.Text))
		case *comment.Code:
			p.writeString(".EX\n%s.EE\n", c.Text)
		case *comment.List:
			for _, item := range c.Items {
				symbol := bullet
				if item.Number != "" {
					symbol = item.Number + "."
				}
				p.writeString(".IP %q\n", symbol)
				p.writeContent(item.Content, depth+1)
			}
		}
	}
}

func (p *Printer) writeString(format string, args ...any) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, format, args...)
}

type HeadingText []comment.Text

func (t HeadingText) String() string {
	var s strings.Builder
	for _, v := range t {
		switch v := v.(type) {
		case comment.Plain:
			s.WriteString(strings.ToUpper(Text(v).String()))
		case comment.Italic:
			s.WriteString(strings.ToUpper(Text(v).String()))
		case *comment.Link:
		case *comment.DocLink:
		}
	}
	return s.String()
}

type ParagraphText []comment.Text

func (t ParagraphText) String() string {
	var s strings.Builder
	for _, v := range t {
		switch v := v.(type) {
		case comment.Plain:
			s.WriteString(Text(v).String())
			s.WriteByte('\n')
		case comment.Italic:
			s.WriteString(".I ")
			s.WriteString(Text(v).String())
			s.WriteByte('\n')
		case *comment.Link:
			fmt.Fprintf(&s, ".UR %s\n", v.URL)
			if len(v.Text) > 0 {
				fmt.Fprintf(&s, "%s\n", LinkText(v.Text))
			}
			s.WriteString(".UE ")
		case *comment.DocLink:
		}
	}
	return s.String()
}

type LinkText []comment.Text

func (t LinkText) String() string {
	var s strings.Builder
	for _, v := range t {
		switch v := v.(type) {
		case comment.Plain:
			s.WriteString(Text(v).String())
		case comment.Italic:
			s.WriteString(Text(v).String())
		case *comment.Link:
		case *comment.DocLink:
		}
	}
	return s.String()
}

type text[T ~string] struct {
	s T
}

func Text[T ~string](s T) text[T] {
	return text[T]{s}
}

var escaper = strings.NewReplacer(
	"'", `\'`,
	"`", "\\`",
	"-", `\-`,
	`"`, `\"`,
	"%", `\%`,
)

func (t text[T]) String() string {
	return escaper.Replace(string(t.s))
}
