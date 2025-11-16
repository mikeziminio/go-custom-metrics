package server

import "github.com/mikeziminio/go-custom-metrics/internal/model"

type updateReqSchema struct {
	ID    string           `json:"id"`
	MType model.MetricType `json:"type"`
	Delta *int64           `json:"delta,omitempty"`
	Value *float64         `json:"value,omitempty"`
}

type updatesReqSchema []updateReqSchema

type getReqSchema struct {
	ID    string           `json:"id"`
	MType model.MetricType `json:"type"`
}
