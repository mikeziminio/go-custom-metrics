package helper

import (
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
