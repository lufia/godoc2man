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
	"path"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "usage: %s [pkg ...]\n", filepath.Base(os.Args[0]))
	fmt.Fprint(w, "\noptions:\n")
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
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(c, name)
	if err != nil {
		log.Fatalln(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		log.Fatalln("too many errors")
	}
	for _, pkg := range pkgs {
		p, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkg.ID)
		if err != nil {
			log.Fatalln("parsing documents:", err)
		}
		if pkg.Name != "main" {
			continue
		}

		var flags []*Flag
		for flg := range FindFlags(pkg.TypesInfo, pkg.Fset, pkg.Syntax) {
			flags = append(flags, flg)
		}
		f, err := os.Create(path.Base(pkg.ID) + ".1")
		if err != nil {
			log.Fatalln("failed to create a file:", err)
		}
		var parser comment.Parser
		doc := parser.Parse(p.Doc)
		printer := NewPrinter(pkg.ID, 1, f)
		printer.Command(p, doc, flags)
		if err := printer.Err(); err != nil {
			log.Fatalln(err)
		}
		if err := f.Sync(); err != nil {
			log.Fatalln(err)
		}
		f.Close()
	}
}
