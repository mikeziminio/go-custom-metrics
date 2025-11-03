package helper

import (
	"path/filepath"
	"testing"
)

func NewInt64(t *testing.T, v int64) *int64 {
	t.Helper()
	return &v
}

func NewFloat64(t *testing.T, v float64) *float64 {
	t.Helper()
	return &v
}

func TempFilePath(t *testing.T, name string) string {
	t.Helper()
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, name)
}
