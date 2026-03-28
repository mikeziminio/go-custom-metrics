package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "customlint",
		Doc:  "check for panic and log.Fatal/os.Exit usage",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				isMain := fn.Name.Name == "main" && pass.Pkg.Name() == "main"
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					fnName := getFuncName(call.Fun)
					if fnName == "panic" {
						pass.Report(analysis.Diagnostic{
							Pos:     call.Pos(),
							Message: "panic not allowed",
						})
						return true
					}
					if !isMain && (fnName == "log.Fatal" || fnName == "os.Exit") {
						pass.Report(analysis.Diagnostic{
							Pos:     call.Pos(),
							Message: "log.Fatal/os.Exit not allowed outside main",
						})
						return true
					}
					return true
				})
			}
		}
	}
	return nil, nil
}

func getFuncName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		if p, ok := v.X.(*ast.Ident); ok {
			return p.Name + "." + v.Sel.Name
		}
	}
	return ""
}

func main() { singlechecker.Main(NewAnalyzer()) }
