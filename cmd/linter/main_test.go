package main

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata, _ := filepath.Abs(filepath.Join("..", "..", "testdata", "linter"))
	analysistest.Run(t, testdata, NewAnalyzer(), "a", "b", "c")
}
