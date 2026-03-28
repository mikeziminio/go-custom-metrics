package main

import (
	"go/ast"
	"go/types"

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
					fnName := getFuncName(pass, call.Fun)
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

func getFuncName(pass *analysis.Pass, e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		id, ok := v.X.(*ast.Ident)
		if !ok {
			return v.Sel.Name
		}
		obj := pass.TypesInfo.Uses[id]
		pkgName := ""
		if obj != nil {
			if pn, ok := obj.(*types.PkgName); ok {
				if imported := pn.Imported(); imported != nil {
					pkgName = imported.Name()
				}
			}
		}
		if pkgName != "" {
			return pkgName + "." + v.Sel.Name
		}
		return id.Name + "." + v.Sel.Name
	}
	return ""
}

func main() {
	singlechecker.Main(NewAnalyzer())
}
