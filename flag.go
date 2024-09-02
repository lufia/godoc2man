package main

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/types/typeutil"
)

type Flag struct {
	Name        string
	Placeholder string
	Usage       string
}

func FindFlags(info *types.Info, fset *token.FileSet, files []*ast.File) <-chan *Flag {
	c := make(chan *Flag)
	go func() {
		for _, f := range files {
			for _, obj := range f.Scope.Objects {
				p, ok := obj.Decl.(ast.Node)
				if !ok {
					continue
				}
				ast.Inspect(p, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					if flg := flagFunc(info, call); flg != nil {
						name, usage := flag.UnquoteUsage(flg)
						c <- &Flag{
							Name:        flg.Name,
							Placeholder: name,
							Usage:       usage,
						}
					}
					return true
				})
			}
		}
		close(c)
	}()
	return c
}

var (
	basicFlags = []string{
		"Bool",
		"Duration",
		"Float64",
		"Int",
		"Int64",
		"String",
		"Uint",
		"Uint64",
	}
	varFlags  = variants(basicFlags, "Var")
	funcFlags = variants(basicFlags, "Func")
)

func variants(flags []string, suffix string) []string {
	a := make([]string, len(flags))
	for i, s := range flags {
		a[i] = s + suffix
	}
	return a
}

func flagFunc(info *types.Info, call *ast.CallExpr) *flag.Flag {
	obj := typeutil.Callee(info, call)
	if obj == nil || obj.Pkg() == nil {
		return nil
	}
	if obj.Pkg().Path() != "flag" {
		return nil
	}
	switch {
	default:
		return nil
	case slices.Contains(basicFlags, obj.Name()) && len(call.Args) == 3:
		return &flag.Flag{
			Name:     exprStr(call.Args[0]),
			Usage:    exprStr(call.Args[2]),
			DefValue: exprStr(call.Args[1]),
		}
	case slices.Contains(varFlags, obj.Name()) && len(call.Args) == 4:
		return &flag.Flag{
			Name:     exprStr(call.Args[1]),
			Usage:    exprStr(call.Args[3]),
			DefValue: exprStr(call.Args[2]),
		}
	case slices.Contains(funcFlags, obj.Name()) && len(call.Args) == 3:
		return &flag.Flag{
			Name:  exprStr(call.Args[0]),
			Usage: exprStr(call.Args[1]),
		}
	}
}

func exprStr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.BasicLit:
		n := len(t.Value)
		if n >= 2 && t.Value[0] == '"' && t.Value[n-1] == '"' {
			return t.Value[1 : n-1]
		}
		return t.Value
	}
	return ""
}
