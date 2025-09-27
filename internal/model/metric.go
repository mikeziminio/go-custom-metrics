package model

import (
	"errors"
	"fmt"
)

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

func NewMetricTypeFromString(s string) (MetricType, error) {
	switch MetricType(s) {
	case Counter, Gauge:
		return MetricType(s), nil
	default:
		return MetricType(""), fmt.Errorf("incorrect metric type")
	}
}

type Metric struct {
	ID    string     `json:"id"`
	MType MetricType `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
	Hash  string     `json:"hash,omitempty"`
}

var MetricNotFoundErr = errors.New("metric not found")
