package model

import (
	"errors"
)

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

// NewMetricTypeFromString converts a string to a MetricType.
//
// Returns an error if the string does not match "counter" or "gauge".
func NewMetricTypeFromString(s string) (MetricType, error) {
	switch MetricType(s) {
	case Counter, Gauge:
		return MetricType(s), nil
	default:
		return MetricType(""), ErrIncorrectMetricType
	}
}

type Metric struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
}

var (
	// ErrMetricNotFound is returned when a metric is not found.
	ErrMetricNotFound = errors.New("metric not found")
	// ErrIncorrectMetricType is returned when an invalid metric type is provided.
	ErrIncorrectMetricType = errors.New("incorrect metric type")
)
