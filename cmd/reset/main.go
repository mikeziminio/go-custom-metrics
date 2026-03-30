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
	Name      string
	Fields    []fieldInfo
	Doc       string
	HasReset  bool
	types     map[string]struct{}
	typeDecls map[string]string
}

type fieldInfo struct {
	Name        string
	Type        string
	PointTo     string
	ElementType string
	BaseType    string
	IsPtr       bool
	IsSlice     bool
	IsArray     bool
	IsMap       bool
	IsStruct    bool
}

var ignoredDirs = map[string]struct{}{
	".git":       {},
	".zed":       {},
	".aidex":     {},
	"migrations": {},
	"bin":        {},
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

		genFilePath := filepath.Join(pkgPath, "reset.gen.go")
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

	// Первый проход: собираем все объявления типов в пакете
	allTypeDecls := make(map[string]string)
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

				allTypeDecls[typeSpec.Name.Name] = astExprToString(typeSpec.Type)
			}
		}
	}

	// Второй проход: извлекаем структуры с комментарием generate:reset
	var structs []structInfo
	for _, file := range files {
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Проверяем комментарии на уровне файла ( комментарии в начале файла перед package)
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

				// Проверяем документацию объявления и тип
				declHasReset := hasGenerateResetComment(genDecl.Doc)
				typeHasReset := hasGenerateResetComment(typeSpec.Doc)

				// Также проверяем, есть ли комментарий с reset на уровне файла
				if !fileHasResetComment && !declHasReset && !typeHasReset {
					continue
				}

				structs = append(structs, extractStructInfo(typeSpec.Name.Name, structType, allTypeDecls))
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

func extractStructInfo(name string, structType *ast.StructType, typeDecls map[string]string) structInfo {
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
				fieldInfo.BaseType = getBaseType(fieldInfo.PointTo, typeDecls)
			case *ast.ArrayType:
				if t.Len == nil {
					fieldInfo.IsSlice = true
					fieldInfo.ElementType = astExprToString(t.Elt)
					fieldInfo.BaseType = getBaseType(fieldInfo.ElementType, typeDecls)
				} else {
					fieldInfo.IsArray = true
					fieldInfo.ElementType = astExprToString(t.Elt)
					fieldInfo.BaseType = getBaseType(fieldInfo.ElementType, typeDecls)
				}
			case *ast.MapType:
				fieldInfo.IsMap = true
				fieldInfo.ElementType = astExprToString(t.Value)
				fieldInfo.BaseType = getBaseType(fieldInfo.ElementType, typeDecls)
			case *ast.StructType:
				fieldInfo.IsStruct = true
				fieldInfo.BaseType = getBaseType(fieldInfo.Type, typeDecls)
			default:
				fieldInfo.BaseType = getBaseType(fieldInfo.Type, typeDecls)
			}

			fields = append(fields, fieldInfo)
		}
	}

	return structInfo{
		Name:      name,
		Fields:    fields,
		HasReset:  false,
		types:     map[string]struct{}{},
		typeDecls: typeDecls,
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

{{range .ResetMethods }}{{.}}{{end}}`

var resetFileTemplate = template.Must(template.New("").Parse(resetFileTemplateSrc))

func generateResetFile(pkgName string, structs []structInfo) (string, error) {
	resetMethods := make([]string, 0, len(structs))
	for _, s := range structs {
		resetMethod, err := generateResetMethod(s)
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

{{range .ResetFields }}	{{.}}
{{end}}}

`

var resetMethodTemplate = template.Must(template.New("").Parse(resetMethodTemplateSrc))

func generateResetMethod(s structInfo) (string, error) {
	var resetFields = make([]string, 0, len(s.Fields))
	for _, f := range s.Fields {
		resetField, err := generateResetField(f, s.types, s.typeDecls)
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

const (
	resetStructPtrTemplateSrc = `if rs.{{.FieldName}} != nil {
	if resetter, ok := any(rs.{{.FieldName}}).(interface{ Reset() }); ok {
		resetter.Reset()
	}
}`
	resetPtrTemplateSrc = `if rs.{{.FieldName}} != nil {
	*rs.{{.FieldName}} = {{.ZeroValue}}
}`
	resetStructTemplateSrc = `if resetter, ok := any(rs.{{.FieldName}}).(interface{ Reset() }); ok {
	resetter.Reset()
}`
	resetSliceTemplateSrc = `rs.{{.FieldName}} = rs.{{.FieldName}}[:0]`
	resetArrayTemplateSrc = `rs.{{.FieldName}} = {{.FieldName}}{}`
	resetMapTemplateSrc   = `clear(rs.{{.FieldName}})`
	resetBaseTemplateSrc  = `rs.{{.FieldName}} = {{.ZeroValue}}`
)

var (
	resetStructPtrTemplate = template.Must(template.New("").Parse(resetStructPtrTemplateSrc))
	resetPtrTemplate       = template.Must(template.New("").Parse(resetPtrTemplateSrc))
	resetStructTemplate    = template.Must(template.New("").Parse(resetStructTemplateSrc))
	resetSliceTemplate     = template.Must(template.New("").Parse(resetSliceTemplateSrc))
	resetArrayTemplate     = template.Must(template.New("").Parse(resetArrayTemplateSrc))
	resetMapTemplate       = template.Must(template.New("").Parse(resetMapTemplateSrc))
	resetBaseTemplate      = template.Must(template.New("").Parse(resetBaseTemplateSrc))
)

func generateResetField(f fieldInfo, structTypes map[string]struct{}, typeDecls map[string]string) (string, error) {
	var sb strings.Builder

	data := struct {
		FieldName string
		ZeroValue string
	}{
		FieldName: f.Name,
		ZeroValue: getZeroValueByBaseType(f.BaseType),
	}

	var err error
	if f.IsPtr {
		if isStructType(f.PointTo, structTypes, typeDecls) {
			err = resetStructPtrTemplate.Execute(&sb, data)
		} else {
			err = resetPtrTemplate.Execute(&sb, data)
		}
	} else if f.IsSlice {
		err = resetSliceTemplate.Execute(&sb, data)
	} else if f.IsArray {
		err = resetArrayTemplate.Execute(&sb, data)
	} else if f.IsMap {
		err = resetMapTemplate.Execute(&sb, data)
	} else if f.IsStruct && isStructType(f.Name, structTypes, typeDecls) {
		err = resetStructTemplate.Execute(&sb, data)
	} else {
		err = resetBaseTemplate.Execute(&sb, data)
	}
	if err != nil {
		return "", err
	}

	sb.WriteString("\n")
	return sb.String(), nil
}

func isStructType(typ string, structTypes map[string]struct{}, typeDecls map[string]string) bool {
	// Check if it's directly in structTypes
	if _, ok := structTypes[typ]; ok {
		return true
	}
	// Check if typ itself is "struct{}"
	if typ == "struct{}" {
		return true
	}
	// For named types, check if typeDecls exists and whether it resolves to struct{}
	// Note: this implementation does NOT recursively resolve type aliases
	if typeDecls != nil {
		if resolved, ok := typeDecls[typ]; ok {
			return resolved == "struct{}"
		}
	}
	return false
}

func getBaseType(typ string, typeDecls map[string]string) string {
	if typ == "" {
		return ""
	}

	// Handle pointer types
	if strings.HasPrefix(typ, "*") {
		base := typ[1:]
		// Resolve through typeDecls if available
		if typeDecls != nil {
			if resolved, ok := typeDecls[base]; ok {
				return getBaseType(resolved, nil)
			}
		}
		return base
	}

	// Handle array types [N]T which should not be reduced to base type
	// since arrays need special reset handling
	if strings.HasPrefix(typ, "[") {
		return typ
	}

	// Handle slice types []T
	if strings.HasPrefix(typ, "[]") {
		base := typ[2:]
		// Resolve through typeDecls if available
		if typeDecls != nil {
			if resolved, ok := typeDecls[base]; ok {
				return getBaseType(resolved, nil)
			}
		}
		return base
	}

	// Handle map types
	if strings.HasPrefix(typ, "map[") {
		// Extract the value type from map[key]value
		closingBracket := strings.Index(typ, "]")
		if closingBracket > 0 {
			base := typ[closingBracket+1:]
			// Resolve through typeDecls if available
			if typeDecls != nil {
				if resolved, ok := typeDecls[base]; ok {
					return getBaseType(resolved, nil)
				}
			}
			return base
		}
		return typ
	}

	// Check if it's a named type that should be resolved
	if typeDecls != nil {
		if resolved, ok := typeDecls[typ]; ok {
			return getBaseType(resolved, nil)
		}
	}

	switch {
	case typ == "int", typ == "int8", typ == "int16", typ == "int32", typ == "int64":
		return "int"
	case typ == "uint", typ == "uint8", typ == "uint16", typ == "uint32", typ == "uint64":
		return "uint"
	case typ == "float32", typ == "float64":
		return "float"
	case typ == "bool":
		return "bool"
	case typ == "string":
		return "string"
	case typ == "complex64", typ == "complex128":
		return "complex"
	case typ == "rune":
		return "rune"
	case typ == "byte":
		return "byte"
	case typ == "uintptr":
		return "uintptr"
	case typ == "struct{}":
		return "struct{}"
	default:
		return ""
	}
}

func getZeroValueByBaseType(baseType string) string {
	switch baseType {
	case "int":
		return "0"
	case "uint":
		return "0"
	case "float":
		return "0"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "complex":
		return "0 + 0i"
	case "rune":
		return "0"
	case "byte":
		return "0"
	case "uintptr":
		return "0"
	case "struct{}":
		return "{}"
	default:
		return "nil"
	}
}

func getZeroValueWithDecls(typ string, typeDecls map[string]string) string {
	if typ == "" {
		return "nil"
	}

	// For pointer, slice, and map types, return nil
	if strings.HasPrefix(typ, "*") || strings.HasPrefix(typ, "[]") || strings.HasPrefix(typ, "map[") {
		return "nil"
	}

	// For named types, get the base type and use getZeroValueByBaseType
	baseType := getBaseType(typ, typeDecls)
	return getZeroValueByBaseType(baseType)
}

func getZeroValue(typ string) string {
	if typ == "" {
		return "nil"
	}

	// For pointer, slice, and map types, return nil
	if strings.HasPrefix(typ, "*") || strings.HasPrefix(typ, "[]") || strings.HasPrefix(typ, "map[") {
		return "nil"
	}

	// For array types [N]T, return zero value for that array type
	if strings.HasPrefix(typ, "[") {
		closingBracket := strings.Index(typ, "]")
		if closingBracket > 0 {
			elemType := typ[closingBracket+1:]
			elemZero := getZeroValueByBaseType(getBaseType(elemType, nil))
			// Extract array size from [N]T
			arrayContent := typ[1:closingBracket]
			// Parse the size (N) - handle both [5]int and [...]int
			size := 0
			if strings.Contains(arrayContent, "...") {
				// [...]T - variable size, just use 2 elements for zero value
				size = 2
			} else {
				// [N]T - fixed size
				sizeStr := strings.TrimSpace(arrayContent)
				fmt.Sscanf(sizeStr, "%d", &size)
			}
			if size == 0 {
				return typ + "{}"
			}
			// Build zero value with correct number of elements
			elements := make([]string, size)
			for i := 0; i < size; i++ {
				elements[i] = elemZero
			}
			return typ + "{" + strings.Join(elements, ",") + "}"
		}
		return typ + "{}"
	}

	// For named types, get the base type and use getZeroValueByBaseType
	baseType := getBaseType(typ, nil)
	return getZeroValueByBaseType(baseType)
}

// + has unit test
// ignoreDir checks if the directory should be ignored during the walk.
func ignoreDir(name string) bool {
	_, ok := ignoredDirs[name]
	return ok
}
