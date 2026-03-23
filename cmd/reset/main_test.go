package main

import (
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
							{Name: "Name", Type: "string"},
							{Name: "Age", Type: "int"},
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
							{Name: "Name", Type: "string"},
							{Name: "Age", Type: "int"},
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
							{Name: "Name", Type: "string"},
						},
					},
					{
						Name: "Post",
						Fields: []fieldInfo{
							{Name: "Title", Type: "string"},
							{Name: "Body", Type: "string"},
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
							{Name: "Name", Type: "string"},
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
							{Name: "Name", Type: "string"},
						},
					},
					{
						Name: "Post",
						Fields: []fieldInfo{
							{Name: "Title", Type: "string"},
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
							{Name: "Name", Type: "string"},
							{Name: "Age", Type: "int"},
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
							{Name: "Ptr", Type: "*string", PointTo: "string", IsPtr: true},
							{Name: "Slice", Type: "[]int", ElementType: "int", IsSlice: true},
							{Name: "Map", Type: "map[string]int", ElementType: "int", IsMap: true},
							{Name: "Struct", Type: "struct{}", IsStruct: true},
							{Name: "Named", Type: "User"},
							{Name: "Pointer", Type: "*User", PointTo: "User", IsPtr: true},
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
							{Name: "Name", Type: "string"},
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
							{Name: "Name", Type: "string"},
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
							{Name: "Name", Type: "string"},
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
							{Name: "ID", Type: "int"},
							{Name: "Name", Type: "string"},
							{Name: "Active", Type: "bool"},
							{Name: "Score", Type: "float64"},
							{Name: "CreatedAt", Type: "int64"},
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
							{Name: "Name", Type: "string"},
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
