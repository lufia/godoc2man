// godoc2man generates man pages.
//
// # SYNPOSIS
//
//	godoc2man [pkg ...]
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/doc/comment"
	"log"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "usage: %s [pkg ...]\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("godoc2man: ")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		Run(".")
	} else {
		for _, name := range flag.Args() {
			Run(name)
		}
	}
}

func Run(name string) {
	c := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax}
	pkgs, err := packages.Load(c, name)
	if err != nil {
		log.Fatalln(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		log.Fatalln("too many errors")
	}
	for _, pkg := range pkgs {
		var s strings.Builder
		for _, f := range pkg.Syntax {
			if err := collectComment(&s, f.Doc); err != nil {
				log.Fatalln("failed to read comments:", err)
			}
		}
		var p comment.Parser
		doc := p.Parse(s.String())
		if err := writeDoc(os.Stdout, doc); err != nil {
			log.Fatalln("failed to output doc:", err)
		}
	}
}

func collectComment(s *strings.Builder, doc *ast.CommentGroup) error {
	if doc == nil {
		return nil
	}
	for _, c := range doc.List {
		if strings.HasPrefix(c.Text, "//go:") {
			continue
		}
		t := strings.TrimPrefix(c.Text, "//")
		if len(t) > 0 && t[0] == ' ' {
			t = t[1:]
		}
		if _, err := s.WriteString(t); err != nil {
			return err
		}
		if err := s.WriteByte('\n'); err != nil {
			return err
		}
	}
	return nil
}

func writeDoc(w io.Writer, doc *comment.Doc) error {
	fmt.Fprintf(w, ".TH name 1\n")
	for _, b := range doc.Content {
		switch e := b.(type) {
		case *comment.Code:
			_, err := fmt.Fprintf(w, "%s\n", e.Text)
			if err != nil {
				return err
			}
		case *comment.Heading:
			fmt.Fprint(w, ".SH ")
			writeText(w, e.Text)
		case *comment.List:
		case *comment.Paragraph:
			fmt.Fprintln(w, ".PP")
			writeText(w, e.Text)
		}
	}
	return nil
}

func writeText(w io.Writer, s []comment.Text) error {
	for _, t := range s {
		switch t := t.(type) {
		case comment.Plain:
			_, err := fmt.Fprintf(w, "%s\n", t)
			return err
		case comment.Italic:
			_, err := fmt.Fprintf(w, ".I %s\n", t)
			return err
		case *comment.Link:
		case *comment.DocLink:
		}
	}
	return nil
}
