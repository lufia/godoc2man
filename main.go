// godoc2man generates man pages.
//
// # SYNPOSIS
//
//	godoc2man [pkg ...]
package main

import (
	"flag"
	"fmt"
	"go/doc"
	"go/doc/comment"
	"log"
	"os"
	"path/filepath"

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
	c := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(c, name)
	if err != nil {
		log.Fatalln(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		log.Fatalln("too many errors")
	}
	for _, pkg := range pkgs {
		if pkg.Name != "main" {
			continue
		}
		p, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkg.ID)
		if err != nil {
			log.Fatalln("parsing documents:", err)
		}
		var parser comment.Parser
		doc := parser.Parse(p.Doc)
		printer := NewPrinter(pkg.ID, 1, os.Stdout)
		printer.Command(p, doc)
		if err := printer.Err(); err != nil {
			log.Fatalln(err)
		}
	}
}
