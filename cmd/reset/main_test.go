package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go/ast"
)

// Helper function to compare struct slices in a deterministic way
func assertStructSlicesEqual(t *testing.T, expected, actual []structInfo) {
	assert.Equal(t, len(expected), len(actual), "slice lengths should match")

	// Create a map for easier lookup
	actualMap := make(map[string]structInfo)
	for _, s := range actual {
		actualMap[s.Name] = s
	}

	for _, exp := range expected {
		act, exists := actualMap[exp.Name]
		assert.True(t, exists, "struct %s should exist", exp.Name)
		assert.Equal(t, exp.Name, act.Name)
		assert.Equal(t, exp.Fields, act.Fields)
		assert.Equal(t, exp.Doc, act.Doc)
	}
}

func TestIgnoreDir(t *testing.T) {
	testCases := []struct {
		dirName string
		want    bool
	}{
		{".git", true},
		{".zed", true},
		{".aidex", true},
		{"migrations", true},
		{"bin", true},
		{"cmd", false},
		{"internal", false},
		{"docs", false},
		{"test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.dirName, func(t *testing.T) {
			got := ignoreDir(tc.dirName)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHasGenerateResetComment(t *testing.T) {
	testCases := []struct {
		name     string
		comments *ast.CommentGroup
		want     bool
	}{
		{
			name:     "nil comment group",
			comments: nil,
			want:     false,
		},
		{
			name: "comment with generate:reset",
			comments: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// generate:reset"},
			}},
			want: true,
		},
		{
			name: "comment without generate:reset",
			comments: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// some comment"},
			}},
			want: false,
		},
		{
			name: "multiple comments with generate:reset",
			comments: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// some comment"},
				{Text: "// generate:reset"},
				{Text: "// another comment"},
			}},
			want: true,
		},
		{
			name: "multiple comments without generate:reset",
			comments: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// comment 1"},
				{Text: "// comment 2"},
			}},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := hasGenerateResetComment(tc.comments)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFindPackagesWithResetComments(t *testing.T) {
	testCases := []struct {
		name    string
		testDir string
		expect  map[string][]structInfo
	}{
		{
			name:    "single struct with generate:reset in comment group",
			testDir: "single_struct_comment_group",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "single_struct_comment_group"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
							{Name: "Age", Type: "int", BaseType: "int"},
						},
					},
				},
			},
		},
		{
			name:    "single struct with generate:reset in doc comment",
			testDir: "single_struct_doc_comment",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "single_struct_doc_comment"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
							{Name: "Age", Type: "int", BaseType: "int"},
						},
					},
				},
			},
		},
		{
			name:    "multiple structs in one file",
			testDir: "multiple_structs_one_file",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "multiple_structs_one_file"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
					{
						Name: "Post",
						Fields: []fieldInfo{
							{Name: "Title", Type: "string", BaseType: "string"},
							{Name: "Body", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "struct without generate:reset should be skipped",
			testDir: "struct_without_comment",
			expect:  map[string][]structInfo{},
		},
		{
			name:    "mixed: some with comment, some without",
			testDir: "mixed_with_without",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "mixed_with_without"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "multiple files in one package",
			testDir: "multiple_files_package",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "multiple_files_package"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
					{
						Name: "Post",
						Fields: []fieldInfo{
							{Name: "Title", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "nested directories",
			testDir: "nested_dirs",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "nested_dirs", "models"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
							{Name: "Age", Type: "int", BaseType: "int"},
						},
					},
				},
			},
		},
		{
			name:    "struct with various field types",
			testDir: "various_field_types",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "various_field_types"): {
					{
						Name: "Complex",
						Fields: []fieldInfo{
							{Name: "Ptr", Type: "*string", PointTo: "string", BaseType: "string", IsPtr: true},
							{Name: "Slice", Type: "[]int", ElementType: "int", BaseType: "int", IsSlice: true},
							{Name: "Map", Type: "map[string]int", ElementType: "int", BaseType: "int", IsMap: true},
							{Name: "Struct", Type: "struct{}", BaseType: "struct{}", IsStruct: true},
							{Name: "Named", Type: "User", BaseType: ""},
							{Name: "Pointer", Type: "*User", PointTo: "User", BaseType: "", IsPtr: true},
						},
					},
				},
			},
		},
		{
			name:    "file with multiple comments in comment group",
			testDir: "multiple_comments_group",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "multiple_comments_group"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "file with syntax error should be skipped",
			testDir: "syntax_error_file",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "syntax_error_file"): {
					{
						Name: "Valid",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "only test files should be ignored",
			testDir: "test_files_ignored",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "test_files_ignored"): {
					{
						Name: "MainStruct",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
		{
			name:    "struct fields with different types",
			testDir: "fields_different_types",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "fields_different_types"): {
					{
						Name: "AllFields",
						Fields: []fieldInfo{
							{Name: "ID", Type: "int", BaseType: "int"},
							{Name: "Name", Type: "string", BaseType: "string"},
							{Name: "Active", Type: "bool", BaseType: "bool"},
							{Name: "Score", Type: "float64", BaseType: "float"},
							{Name: "CreatedAt", Type: "int64", BaseType: "int"},
						},
					},
				},
			},
		},
		{
			name:    "ignored directory .git should be skipped",
			testDir: "ignored_git_dir",
			expect: map[string][]structInfo{
				filepath.Join("testdata", "reset", "ignored_git_dir", "models"): {
					{
						Name: "User",
						Fields: []fieldInfo{
							{Name: "Name", Type: "string", BaseType: "string"},
						},
					},
				},
			},
		},
	}

	basePath := "testdata/reset"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testPath := filepath.Join(basePath, tc.testDir)

			result, err := findPackagesWithResetComments(testPath)
			assert.NoError(t, err)

			// Compare maps - for each package path, compare struct slices
			assert.Equal(t, len(tc.expect), len(result), "number of packages should match")
			for pkgPath, expectedStructs := range tc.expect {
				actualStructs, exists := result[pkgPath]
				assert.True(t, exists, "package %s should exist", pkgPath)
				assertStructSlicesEqual(t, expectedStructs, actualStructs)
			}
		})
	}
}

func TestPackageName(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "regular package",
			filePath: "/path/to/pkg/file.go",
			want:     "pkg",
		},
		{
			name:     "internal package at base",
			filePath: "/path/to/internal/file.go",
			want:     "internal",
		},
		{
			name:     "nested path",
			filePath: "/path/to/some/deep/nested/pkg/file.go",
			want:     "pkg",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := packageName(tc.filePath)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGenerateResetMethod(t *testing.T) {
	testCases := []struct {
		name         string
		structInfo   structInfo
		structTypes  map[string]struct{}
		wantContains []string
	}{
		{
			name: "simple struct with basic fields",
			structInfo: structInfo{
				Name: "Simple",
				Fields: []fieldInfo{
					{Name: "ID", Type: "int", BaseType: "int"},
					{Name: "Name", Type: "string", BaseType: "string"},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *Simple) Reset()",
				"rs.ID = 0",
				"rs.Name = \"\"",
			},
		},
		{
			name: "struct with pointer field",
			structInfo: structInfo{
				Name: "WithPtr",
				Fields: []fieldInfo{
					{Name: "Value", Type: "*int", PointTo: "int", BaseType: "int", IsPtr: true},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *WithPtr) Reset()",
				"if rs.Value != nil",
				"*rs.Value = 0",
			},
		},
		{
			name: "struct with slice field",
			structInfo: structInfo{
				Name: "WithSlice",
				Fields: []fieldInfo{
					{Name: "Items", Type: "[]int", ElementType: "int", BaseType: "int", IsSlice: true},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *WithSlice) Reset()",
				"rs.Items = rs.Items[:0]",
			},
		},
		{
			name: "struct with map field",
			structInfo: structInfo{
				Name: "WithMap",
				Fields: []fieldInfo{
					{Name: "Data", Type: "map[string]int", ElementType: "int", BaseType: "int", IsMap: true},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *WithMap) Reset()",
				"clear(rs.Data)",
			},
		},
		{
			name: "struct with multiple fields including all types",
			structInfo: structInfo{
				Name: "MultiFields",
				Fields: []fieldInfo{
					{Name: "ID", Type: "int", BaseType: "int"},
					{Name: "Name", Type: "string", BaseType: "string"},
					{Name: "Ptr", Type: "*string", PointTo: "string", BaseType: "string", IsPtr: true},
					{Name: "Slice", Type: "[]int", ElementType: "int", BaseType: "int", IsSlice: true},
					{Name: "Map", Type: "map[string]int", ElementType: "int", BaseType: "int", IsMap: true},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *MultiFields) Reset()",
				"rs.ID = 0",
				"rs.Name = \"\"",
				"if rs.Ptr != nil",
				"*rs.Ptr = \"\"",
				"rs.Slice = rs.Slice[:0]",
				"clear(rs.Map)",
			},
		},
		{
			name: "struct with float field",
			structInfo: structInfo{
				Name: "WithFloat",
				Fields: []fieldInfo{
					{Name: "Score", Type: "float64", BaseType: "float"},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *WithFloat) Reset()",
				"rs.Score = 0",
			},
		},
		{
			name: "struct with bool field",
			structInfo: structInfo{
				Name: "WithBool",
				Fields: []fieldInfo{
					{Name: "Active", Type: "bool", BaseType: "bool"},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{},
			wantContains: []string{
				"func (rs *WithBool) Reset()",
				"rs.Active = false",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := generateResetMethod(tc.structInfo, tc.structTypes)
			assert.NoError(t, err)

			for _, want := range tc.wantContains {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestGenerateResetMethodWithStructTypes(t *testing.T) {
	testCases := []struct {
		name         string
		structInfo   structInfo
		structTypes  map[string]struct{}
		wantContains string
	}{
		{
			name: "pointer to struct with reset",
			structInfo: structInfo{
				Name: "WithStructPtr",
				Fields: []fieldInfo{
					{Name: "Config", Type: "*AppConfig", PointTo: "AppConfig", BaseType: "AppConfig", IsPtr: true},
				},
				HasReset: false,
			},
			structTypes: map[string]struct{}{
				"AppConfig": {},
			},
			wantContains: "if resetter, ok := (*rs.Config).(interface{ Reset() }); ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := generateResetMethod(tc.structInfo, tc.structTypes)
			assert.NoError(t, err)
			assert.Contains(t, got, tc.wantContains)
		})
	}
}

func TestAstExprToString(t *testing.T) {
	testCases := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "simple identifier",
			expr: &ast.Ident{Name: "string"},
			want: "string",
		},
		{
			name: "pointer type",
			expr: &ast.StarExpr{X: &ast.Ident{Name: "User"}},
			want: "*User",
		},
		{
			name: "slice type",
			expr: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}},
			want: "[]int",
		},
		{
			name: "map type",
			expr: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			want: "map[string]int",
		},
		{
			name: "struct type",
			expr: &ast.StructType{},
			want: "struct{}",
		},
		{
			name: "selector expression",
			expr: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Print"}},
			want: "fmt.Print",
		},
		{
			name: "func type",
			expr: &ast.FuncType{},
			want: "func()",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := astExprToString(tc.expr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetZeroValue(t *testing.T) {
	testCases := []struct {
		name string
		typ  string
		want string
	}{
		{
			name: "int type",
			typ:  "int",
			want: "0",
		},
		{
			name: "int8 type",
			typ:  "int8",
			want: "0",
		},
		{
			name: "int16 type",
			typ:  "int16",
			want: "0",
		},
		{
			name: "int32 type",
			typ:  "int32",
			want: "0",
		},
		{
			name: "int64 type",
			typ:  "int64",
			want: "0",
		},
		{
			name: "uint type",
			typ:  "uint",
			want: "0",
		},
		{
			name: "uint8 type",
			typ:  "uint8",
			want: "0",
		},
		{
			name: "uint16 type",
			typ:  "uint16",
			want: "0",
		},
		{
			name: "uint32 type",
			typ:  "uint32",
			want: "0",
		},
		{
			name: "uint64 type",
			typ:  "uint64",
			want: "0",
		},
		{
			name: "float32 type",
			typ:  "float32",
			want: "0",
		},
		{
			name: "float64 type",
			typ:  "float64",
			want: "0",
		},
		{
			name: "bool type",
			typ:  "bool",
			want: "false",
		},
		{
			name: "string type",
			typ:  "string",
			want: "\"\"",
		},
		{
			name: "empty type",
			typ:  "",
			want: "nil",
		},
		{
			name: "pointer type",
			typ:  "*string",
			want: "nil",
		},
		{
			name: "slice type",
			typ:  "[]int",
			want: "nil",
		},
		{
			name: "map type",
			typ:  "map[string]int",
			want: "nil",
		},
		{
			name: "struct type",
			typ:  "struct{}",
			want: "{}",
		},
		{
			name: "unknown named type",
			typ:  "CustomType",
			want: "nil",
		},
		{
			name: "array [5]int",
			typ:  "[5]int",
			want: "[5]int{0,0,0,0,0}",
		},
		{
			name: "array [...]string",
			typ:  "[...]string",
			want: "[...]string{\"\",\"\"}",
		},
		{
			name: "array [10]float64",
			typ:  "[10]float64",
			want: "[10]float64{0,0,0,0,0,0,0,0,0,0}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getZeroValue(tc.typ)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsStructType(t *testing.T) {
	testCases := []struct {
		name        string
		typ         string
		structTypes map[string]struct{}
		typeDecls   map[string]string
		want        bool
	}{
		{
			name:        "builtin_int",
			typ:         "int",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false,
		},
		{
			name:        "builtin_string",
			typ:         "string",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false,
		},
		{
			name:        "builtin_bool",
			typ:         "bool",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false,
		},
		{
			name: "known_struct_in_map",
			typ:  "User",
			structTypes: map[string]struct{}{
				"User": {},
			},
			typeDecls: nil,
			want:      true,
		},
		{
			name:        "named_type_with_struct_in_decls",
			typ:         "MyStruct",
			structTypes: map[string]struct{}{},
			typeDecls: map[string]string{
				"MyStruct": "struct{}",
			},
			want: true,
		},
		{
			name:        "pointer_to_struct",
			typ:         "*User",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false, // starts with *, so not a direct struct
		},
		{
			name:        "slice_type",
			typ:         "[]int",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false,
		},
		{
			name:        "map_type",
			typ:         "map[string]int",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false,
		},
		{
			name:        "named_type_with_struct_field",
			typ:         "IntPtr",
			structTypes: map[string]struct{}{},
			typeDecls: map[string]string{
				"IntPtr": "*int",
			},
			want: false,
		},
		{
			name:        "nested_named_type_doesnt_resolve",
			typ:         "MyStructAlias",
			structTypes: map[string]struct{}{},
			typeDecls: map[string]string{
				"MyStructAlias": "MyStruct",
				"MyStruct":      "struct{}",
			},
			want: false, // current implementation doesn't recursively resolve
		},
		{
			name:        "unknown_named_type_assumed_non_struct",
			typ:         "CustomType",
			structTypes: map[string]struct{}{},
			typeDecls:   nil,
			want:        false, // current implementation returns false for unknown types
		},
		{
			name:        "explicit_struct_in_decls",
			typ:         "ExplicitStruct",
			structTypes: map[string]struct{}{},
			typeDecls: map[string]string{
				"ExplicitStruct": "struct{}",
			},
			want: true,
		},
		{
			name:        "named_type_in_struct_types_map",
			typ:         "User",
			structTypes: map[string]struct{}{"User": {}},
			typeDecls:   nil,
			want:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isStructType(tc.typ, tc.structTypes, tc.typeDecls)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetZeroValueWithDecls(t *testing.T) {
	testCases := []struct {
		name      string
		typ       string
		typeDecls map[string]string
		want      string
	}{
		{
			name:      "int_type",
			typ:       "int",
			typeDecls: nil,
			want:      "0",
		},
		{
			name:      "string_type",
			typ:       "string",
			typeDecls: nil,
			want:      "\"\"",
		},
		{
			name:      "bool_type",
			typ:       "bool",
			typeDecls: nil,
			want:      "false",
		},
		{
			name:      "float32_type",
			typ:       "float32",
			typeDecls: nil,
			want:      "0",
		},
		{
			name:      "pointer_type",
			typ:       "*int",
			typeDecls: nil,
			want:      "nil",
		},
		{
			name:      "slice_type",
			typ:       "[]string",
			typeDecls: nil,
			want:      "nil",
		},
		{
			name:      "map_type",
			typ:       "map[string]int",
			typeDecls: nil,
			want:      "nil",
		},
		{
			name:      "struct_type",
			typ:       "struct{}",
			typeDecls: nil,
			want:      "{}",
		},
		{
			name: "named_type_with_decls",
			typ:  "MyInt",
			typeDecls: map[string]string{
				"MyInt": "int",
			},
			want: "0",
		},
		{
			name:      "named_type_without_decls",
			typ:       "CustomType",
			typeDecls: nil,
			want:      "nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getZeroValueWithDecls(tc.typ, tc.typeDecls)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIntegrationGenerateResetFile(t *testing.T) {
	tempDir := t.TempDir()
	testStructs := []structInfo{
		{
			Name: "TestStruct",
			Fields: []fieldInfo{
				{Name: "ID", Type: "int", BaseType: "int"},
				{Name: "Name", Type: "string", BaseType: "string"},
				{Name: "Active", Type: "bool", BaseType: "bool"},
			},
			HasReset: false,
		},
		{
			Name: "ComplexStruct",
			Fields: []fieldInfo{
				{Name: "Ptr", Type: "*string", PointTo: "string", BaseType: "string", IsPtr: true},
				{Name: "Slice", Type: "[]int", ElementType: "int", BaseType: "int", IsSlice: true},
				{Name: "Map", Type: "map[string]int", ElementType: "int", BaseType: "int", IsMap: true},
			},
			HasReset: false,
		},
	}

	filePath := filepath.Join(tempDir, "reset_test.gen.go")
	err := writeResetFile(filePath, testStructs)
	assert.NoError(t, err)

	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "func (rs *TestStruct) Reset()")
	assert.Contains(t, contentStr, "func (rs *ComplexStruct) Reset()")
	assert.Contains(t, contentStr, "rs.ID = 0")
	assert.Contains(t, contentStr, "rs.Name = \"\"")
	assert.Contains(t, contentStr, "rs.Active = false")
	assert.Contains(t, contentStr, "if rs.Ptr != nil")
	assert.Contains(t, contentStr, "rs.Slice = rs.Slice[:0]")
	assert.Contains(t, contentStr, "clear(rs.Map)")
}

func TestGenerateResetFile(t *testing.T) {
	structs := []structInfo{
		{
			Name: "User",
			Fields: []fieldInfo{
				{Name: "ID", Type: "int", BaseType: "int"},
			},
			HasReset: false,
		},
		{
			Name: "Post",
			Fields: []fieldInfo{
				{Name: "Title", Type: "string", BaseType: "string"},
			},
			HasReset: false,
		},
	}

	content, err := generateResetFile("testpkg", structs)
	assert.NoError(t, err)

	assert.Contains(t, content, "package testpkg")
	assert.Contains(t, content, "func (rs *User) Reset()")
	assert.Contains(t, content, "func (rs *Post) Reset()")
	assert.Contains(t, content, "rs.ID = 0")
	assert.Contains(t, content, "rs.Title = \"\"")
}

func TestGenerateResetFileEmptyStructs(t *testing.T) {
	// Test with empty struct slice
	content, err := generateResetFile("empty", []structInfo{})
	assert.NoError(t, err)
	assert.Contains(t, content, "package empty")
	assert.NotContains(t, content, "func (rs *")
}

func TestGenerateResetFileNilStructs(t *testing.T) {
	// Test with nil struct slice
	content, err := generateResetFile("niltest", nil)
	assert.NoError(t, err)
	assert.Contains(t, content, "package niltest")
	assert.NotContains(t, content, "func (rs *")
}

func TestGenerateResetFileNamedTypes(t *testing.T) {
	structs := []structInfo{
		{
			Name: "Config",
			Fields: []fieldInfo{
				{Name: "ID", Type: "MyInt", BaseType: "int"},
				{Name: "Name", Type: "MyString", BaseType: "string"},
				{Name: "Flag", Type: "MyBool", BaseType: "bool"},
			},
			HasReset: false,
		},
	}

	content, err := generateResetFile("pkg", structs)
	assert.NoError(t, err)

	expected := `// Code generated by reset generator; DO NOT EDIT.

package pkg

// Reset Config
func (rs *Config) Reset() {
	if rs == nil {
		return
	}

	rs.ID = 0

	rs.Name = ""

	rs.Flag = false

}

`
	assert.Equal(t, expected, content)
}

func TestGenerateResetFileWithAllFieldTypes(t *testing.T) {
	structs := []structInfo{
		{
			Name: "AllTypes",
			Fields: []fieldInfo{
				{Name: "IntField", Type: "int", BaseType: "int"},
				{Name: "StrField", Type: "string", BaseType: "string"},
				{Name: "BoolField", Type: "bool", BaseType: "bool"},
				{Name: "FloatField", Type: "float64", BaseType: "float"},
				{Name: "PtrInt", Type: "*int", PointTo: "int", BaseType: "int", IsPtr: true},
				{Name: "PtrString", Type: "*string", PointTo: "string", BaseType: "string", IsPtr: true},
				{Name: "SliceInt", Type: "[]int", ElementType: "int", BaseType: "int", IsSlice: true},
				{Name: "SliceStr", Type: "[]string", ElementType: "string", BaseType: "string", IsSlice: true},
				{Name: "MapInt", Type: "map[string]int", ElementType: "int", BaseType: "int", IsMap: true},
			},
			HasReset: false,
		},
	}

	content, err := generateResetFile("pkg", structs)
	assert.NoError(t, err)

	expected := `// Code generated by reset generator; DO NOT EDIT.

package pkg

// Reset AllTypes
func (rs *AllTypes) Reset() {
	if rs == nil {
		return
	}

	rs.IntField = 0

	rs.StrField = ""

	rs.BoolField = false

	rs.FloatField = 0

	if rs.PtrInt != nil {
	*rs.PtrInt = 0
}

	if rs.PtrString != nil {
	*rs.PtrString = ""
}

	rs.SliceInt = rs.SliceInt[:0]

	rs.SliceStr = rs.SliceStr[:0]

	clear(rs.MapInt)

}

`
	assert.Equal(t, expected, content)
}

func TestGenerateResetFileStructWithReset(t *testing.T) {
	structs := []structInfo{
		{
			Name: "Outer",
			Fields: []fieldInfo{
				{Name: "Inner", Type: "InnerStruct", BaseType: "InnerStruct", IsStruct: true},
			},
			HasReset: false,
		},
	}

	content, err := generateResetFile("pkg", structs)
	assert.NoError(t, err)

	expected := `// Code generated by reset generator; DO NOT EDIT.

package pkg

// Reset Outer
func (rs *Outer) Reset() {
	if rs == nil {
		return
	}

	rs.Inner = nil

}

`
	assert.Equal(t, expected, content)
}

func TestGenerateResetFileStructPtrWithReset(t *testing.T) {
	structs := []structInfo{
		{
			Name: "Outer",
			Fields: []fieldInfo{
				{Name: "InnerPtr", Type: "*InnerStruct", PointTo: "InnerStruct", BaseType: "InnerStruct", IsPtr: true},
			},
			HasReset: false,
		},
	}

	content, err := generateResetFile("pkg", structs)
	assert.NoError(t, err)

	expected := `// Code generated by reset generator; DO NOT EDIT.

package pkg

// Reset Outer
func (rs *Outer) Reset() {
	if rs == nil {
		return
	}

	if rs.InnerPtr != nil {
	*rs.InnerPtr = nil
}

}

`
	assert.Equal(t, expected, content)
}
