package server

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func (a *APIServer) Update(res http.ResponseWriter, req *http.Request) {
	a.logger.Info("request start", zap.String("path", req.URL.Path))

	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(req, "metricName")
	var delta *int64
	var value *float64

	val := chi.URLParam(req, "value")
	if metricType == model.Counter {
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		delta = &d
	}

	if metricType == model.Gauge {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		value = &v
	}

	err = a.storage.Update(model.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: delta,
		Value: value,
	})
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
	a.logger.Info("request end")
}

func (a *APIServer) List(res http.ResponseWriter, req *http.Request) {
	a.logger.Info("request start", zap.String("path", req.URL.Path))

	var b bytes.Buffer
	metrics := a.storage.List()
	a.logger.Info("metrics", zap.Int("len", len(metrics)))
	for id, m := range metrics {
		b.WriteString(id)
		switch m.MType {
		case model.Gauge:
			b.WriteString(fmt.Sprintf(" %.5f\n", *m.Value))
		case model.Counter:
			b.WriteString(fmt.Sprintf(" %d\n", *m.Delta))
		}
	}

	_, err := res.Write(b.Bytes())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("failed to write body", zap.Error(err))
	}
}

func (a *APIServer) Get(res http.ResponseWriter, req *http.Request) {
	a.logger.Info("request start", zap.String("path", req.URL.Path))

	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(req, "metricName")

	m, err := a.storage.Get(metricType, metricName)
	if err != nil {
		if errors.Is(err, model.ErrMetricNotFound) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var r string
	if m.MType == model.Gauge {
		r = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	} else {
		r = fmt.Sprintf("%d", *m.Delta)
	}

	_, err = res.Write([]byte(r))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		a.logger.Error("failed to write body", zap.Error(err))
	}
}
