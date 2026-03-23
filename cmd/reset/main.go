package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type structInfo struct {
	Name     string
	Fields   []fieldInfo
	Doc      string
	HasReset bool
	types    map[string]struct{}
}

type fieldInfo struct {
	Name        string
	Type        string
	PointTo     string
	ElementType string
	IsPtr       bool
	IsSlice     bool
	IsMap       bool
	IsStruct    bool
}

var ignoredDirsArr = [...]string{
	".git",
	".zed",
	".aidex",
	"migrations",
	"bin",
}
var ignoredDirs map[string]struct{}

func init() {
	ignoredDirs = make(map[string]struct{})
	for _, id := range ignoredDirsArr {
		ignoredDirs[id] = struct{}{}
	}
}

func main() {
	searchPath := "."
	if len(os.Args) >= 2 {
		searchPath = os.Args[1]
	}

	pkgs, err := findPackagesWithResetComments(searchPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding packages: %v\n", err)
		os.Exit(1)
	}

	for pkgPath, structs := range pkgs {
		if len(structs) == 0 {
			continue
		}

		genFilePath := filepath.Join(filepath.Dir(pkgPath), "reset.gen.go")
		if err := writeResetFile(genFilePath, structs); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", genFilePath, err)
			os.Exit(1)
		}

		fmt.Printf("Generated %s for %d structures\n", genFilePath, len(structs))
	}
}

// + has unit test
func findPackagesWithResetComments(path string) (map[string][]structInfo, error) {
	result := make(map[string][]structInfo)

	err := filepath.Walk(path, func(folderPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if ignoreDir(info.Name()) {
			return filepath.SkipDir
		}

		matches, err := filepath.Glob(filepath.Join(folderPath, "*.go"))
		if err != nil {
			return err
		}

		var goFiles []string
		for _, f := range matches {
			if !strings.HasSuffix(f, "_test.go") {
				goFiles = append(goFiles, f)
			}
		}

		if len(goFiles) == 0 {
			return nil
		}

		return processPackage(folderPath, goFiles, result)
	})

	return result, err
}

// + has unit test
func processPackage(dirPath string, files []string, result map[string][]structInfo) error {
	fset := token.NewFileSet()

	// First pass: collect all struct names in the package
	allStructTypes := make(map[string]struct{})
	for _, file := range files {
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					allStructTypes[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}

	// Second pass: extract structs with generate:reset comment
	var structs []structInfo
	for _, file := range files {
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Check file-level comments (comments at top of file before package)
		fileHasResetComment := hasGenerateResetComment(node.Doc)

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Check declaration doc and type doc
				declHasReset := hasGenerateResetComment(genDecl.Doc)
				typeHasReset := hasGenerateResetComment(typeSpec.Doc)

				// Also check if file-level comment has reset directive
				if !fileHasResetComment && !declHasReset && !typeHasReset {
					continue
				}

				structs = append(structs, extractStructInfo(typeSpec.Name.Name, structType, allStructTypes))
			}
		}
	}

	if len(structs) > 0 {
		result[dirPath] = structs
	}

	return nil
}

// + has unit test
func hasGenerateResetComment(comment *ast.CommentGroup) bool {
	if comment == nil {
		return false
	}

	for _, c := range comment.List {
		if strings.Contains(c.Text, "// generate:reset") {
			return true
		}
	}

	return false
}

func extractStructInfo(name string, structType *ast.StructType, types map[string]struct{}) structInfo {
	var fields []fieldInfo

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue
			}

			fieldInfo := fieldInfo{Name: field.Names[0].Name}

			fieldInfo.Type = astExprToString(field.Type)

			switch t := field.Type.(type) {
			case *ast.StarExpr:
				fieldInfo.IsPtr = true
				fieldInfo.PointTo = astExprToString(t.X)
			case *ast.ArrayType:
				if t.Len == nil {
					fieldInfo.IsSlice = true
					fieldInfo.ElementType = astExprToString(t.Elt)
				}
			case *ast.MapType:
				fieldInfo.IsMap = true
				fieldInfo.ElementType = astExprToString(t.Value)
			case *ast.StructType:
				fieldInfo.IsStruct = true
			}

			fields = append(fields, fieldInfo)
		}
	}

	return structInfo{
		Name:     name,
		Fields:   fields,
		HasReset: false,
		types:    types,
	}
}

func astExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + astExprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + astExprToString(t.Elt)
		}
		return fmt.Sprintf("[%d]%s", t.Len, astExprToString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", astExprToString(t.Key), astExprToString(t.Value))
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func()"
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", astExprToString(t.X), t.Sel.Name)
	default:
		return ""
	}
}

func writeResetFile(filePath string, structs []structInfo) error {
	pkgName := packageName(filePath)
	content, err := generateResetFile(pkgName, structs)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
}

const resetFileTemplateSrc = `// Code generated by reset generator; DO NOT EDIT.

package {{.PackageName}}

{{range .ResetMethods }}
{{.}}
{{end}}
`

var resetFileTemplate = template.Must(template.New("").Parse(resetFileTemplateSrc))

func generateResetFile(pkgName string, structs []structInfo) (string, error) {
	resetMethods := make([]string, 0, len(structs))
	for _, s := range structs {
		resetMethod, err := generateResetMethod(s, s.types)
		if err != nil {
			// todo: добавить логирование (!)
			continue
		}
		resetMethods = append(resetMethods, resetMethod)
	}

	data := struct {
		PackageName  string
		ResetMethods []string
	}{
		PackageName:  pkgName,
		ResetMethods: resetMethods,
	}

	var sb strings.Builder
	if err := resetFileTemplate.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func packageName(filePath string) string {
	dir := filepath.Dir(filePath)
	base := filepath.Base(dir)
	if base == "internal" {
		return "internal"
	}
	return base
}

const resetMethodTemplateSrc = `// Reset {{.StructName}}
func (rs *{{.StructName}}) Reset() {
	if rs == nil {
		return
	}

	{{range .ResetFields }}
	{{.}}
	{{end}}
}
`

var resetMethodTemplate = template.Must(template.New("").Parse(resetMethodTemplateSrc))

func generateResetMethod(s structInfo, structTypes map[string]struct{}) (string, error) {
	var resetFields = make([]string, 0, len(s.Fields))
	for _, f := range s.Fields {
		resetField, err := generateResetField(f, structTypes)
		if err != nil {
			// TODO: логирование или брейк ?
			continue
		}
		resetFields = append(resetFields, resetField)
	}

	data := struct {
		StructName  string
		ResetFields []string
	}{
		StructName:  s.Name,
		ResetFields: resetFields,
	}

	var sb strings.Builder
	if err := resetMethodTemplate.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func generateResetField(f fieldInfo, structTypes map[string]struct{}) (string, error) {
	var sb strings.Builder

	if f.IsPtr {
		if isStructType(f.PointTo, structTypes) {
			fmt.Fprintf(&sb, "if rs.%s != nil {\n", f.Name)
			fmt.Fprintf(&sb, "\tif resetter, ok := (*rs.%s).(interface{ Reset() }); ok {\n", f.Name)
			fmt.Fprintf(&sb, "\t\tresetter.Reset()\n")
			fmt.Fprintf(&sb, "\t}\n")
			fmt.Fprintf(&sb, "}\n")
		} else {
			fmt.Fprintf(&sb, "if rs.%s != nil {\n", f.Name)
			fmt.Fprintf(&sb, "\t*rs.%s = %s\n", f.Name, getZeroValue(f.PointTo))
			fmt.Fprintf(&sb, "}\n")
		}
	} else if f.IsSlice {
		fmt.Fprintf(&sb, "rs.%s = rs.%s[:0]\n", f.Name, f.Name)
	} else if f.IsMap {
		fmt.Fprintf(&sb, "clear(rs.%s)\n", f.Name)
	} else if f.IsStruct && isStructType(f.Name, structTypes) {
		fmt.Fprintf(&sb, "if resetter, ok := rs.%s.(interface{ Reset() }); ok {\n", f.Name)
		fmt.Fprintf(&sb, "\tresetter.Reset()\n")
		fmt.Fprintf(&sb, "}\n")
	} else {
		fmt.Fprintf(&sb, "rs.%s = %s\n", f.Name, getZeroValue(f.Type))
	}

	return sb.String(), nil
}

func isStructType(typ string, structTypes map[string]struct{}) bool {
	_, ok := structTypes[typ]
	return ok
}

func getZeroValue(typ string) string {
	if typ == "" {
		return "nil"
	}

	switch {
	case typ == "int", typ == "int8", typ == "int16", typ == "int32", typ == "int64",
		typ == "uint", typ == "uint8", typ == "uint16", typ == "uint32", typ == "uint64",
		typ == "float32", typ == "float64":
		return "0"
	case typ == "bool":
		return "false"
	case typ == "string":
		return `""`
	case strings.HasPrefix(typ, "*"):
		return "nil"
	case strings.HasPrefix(typ, "[]"):
		return "nil"
	case strings.HasPrefix(typ, "map"):
		return "nil"
	case typ == "struct{}":
		return "{}"
	default:
		return "nil"
	}
}

// + has unit test
// ignoreDir checks if the directory should be ignored during the walk.
func ignoreDir(name string) bool {
	_, ok := ignoredDirs[name]
	return ok
}
