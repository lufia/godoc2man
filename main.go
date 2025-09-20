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
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/lufia/godoc2man/internal/language"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "usage: %s [pkg ...]\n", filepath.Base(os.Args[0]))
	fmt.Fprint(w, "\noptions:\n")
	flag.PrintDefaults()
}

var (
	langFlag = flag.String("lang", "en", "specify the `lang`uage code that is used for GoDoc document")
	flagFlag = flag.String("flag", "none", "generate options section from sources with static analysis; `pkg` is std or none")
	dirFlag  = flag.String("dir", "man", "specify the output `dir`ectory")
)

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

		f, err := outputFile(*dirFlag, pkg.ID, "1")
		if err != nil {
			log.Fatalln("failed to create a file:", err)
		}

		s, err := language.String(*langFlag, p.Doc)
		if err != nil {
			log.Fatalf("failed to transform to language '%s': %v", *langFlag, err)
		}
		var parser comment.Parser
		doc := parser.Parse(s)
		printer := NewPrinter(pkg.Fset, pkg.ID, 1, f)
		if pkg.Name == "main" {
			flags := retrieveFlags(pkg)
			printer.Command(p, doc, flags)
		} else {
			printer.Library(p, doc)
		}
		if err := printer.Err(); err != nil {
			log.Fatalln(err)
		}

		if err := f.Sync(); err != nil {
			log.Fatalln(err)
		}
		f.Close()
	}
}

func outputFile(base, pkgPath, section string) (*os.File, error) {
	dir := filepath.Join(base, "man"+section)
	err := os.MkdirAll(dir, 0755)
	if err != nil && os.IsExist(err) {
		return nil, fmt.Errorf("failed to create %s: %w", dir, err)
	}
	name := strings.ReplaceAll(pkgPath, "/", "-") + "." + section
	file := filepath.Join(dir, name)
	f, err := os.Create(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", file, err)
	}
	return f, nil
}

func retrieveFlags(p *packages.Package) []*Flag {
	var flags []*Flag
	switch *flagFlag {
	default:
		log.Printf("-flag=%s is not supported; ignored\n", *flagFlag)
	case "none":
	case "std":
		for f := range FindFlags(p.TypesInfo, p.Fset, p.Syntax) {
			flags = append(flags, f)
		}
	}
	return flags
}
