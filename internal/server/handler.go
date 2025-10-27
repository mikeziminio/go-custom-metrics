package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func (a *APIServer) Update(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed read request body: %v", err),
			http.StatusBadRequest)
		return
	}

	type reqScheme struct {
		ID    string           `json:"id"`
		MType model.MetricType `json:"type"`
		Delta *int64           `json:"delta,omitempty"`
		Value *float64         `json:"value,omitempty"`
	}
	var data reqScheme

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to validate request body: %v", err),
			http.StatusBadRequest)
		return
	}

	m, err := a.storage.Update(model.Metric{
		ID:    data.ID,
		MType: data.MType,
		Delta: data.Delta,
		Value: data.Value,
	})
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to update metric value %s / %s: %v", data.MType, data.ID, err),
			http.StatusBadRequest)
		return
	}

	resData, err := json.Marshal(m)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to marshal response: %v", err),
			http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(resData)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to write body: %v", err),
			http.StatusInternalServerError,
		)
		return
	}
}

func (a *APIServer) UpdateByParams(res http.ResponseWriter, req *http.Request) {
	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to fetch metric type: %v", err),
			http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(req, "metricName")
	var delta *int64
	var value *float64

	val := chi.URLParam(req, "value")
	if metricType == model.Counter {
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("failed to parse counter value: %v", err),
				http.StatusBadRequest)
			return
		}
		delta = &d
	}

	if metricType == model.Gauge {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("failed to parse gauge value: %v", err),
				http.StatusBadRequest)
			return
		}
		value = &v
	}

	_, err = a.storage.Update(model.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: delta,
		Value: value,
	})
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to update metric value: %v", err),
			http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (a *APIServer) Get(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed read request body: %v", err),
			http.StatusBadRequest)
		return
	}

	type reqScheme struct {
		ID    string           `json:"id"`
		MType model.MetricType `json:"type"`
	}
	var data reqScheme

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to validate request body: %v", err),
			http.StatusBadRequest)
		return
	}

	m, err := a.storage.Get(data.MType, data.ID)
	if err != nil {
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(res, fmt.Sprintf("metric not found %s / %s: %v", data.MType, data.ID, err),
				http.StatusNotFound)
			return
		}
		http.Error(res, fmt.Sprintf("failed to get metric %s / %s: %v", data.MType, data.ID, err),
			http.StatusBadRequest)
		return
	}

	resData, err := json.Marshal(m)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to marshal response: %v", err),
			http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resData)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to write body: %v", err),
			http.StatusInternalServerError)
		return
	}
}

func (a *APIServer) GetByParams(res http.ResponseWriter, req *http.Request) {
	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to fetch metric type: %v", err),
			http.StatusBadRequest)
		return
	}
	metricName := chi.URLParam(req, "metricName")

	m, err := a.storage.Get(metricType, metricName)
	if err != nil {
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(res, fmt.Sprintf("metric not found %s / %s: %v", metricType, metricName, err),
				http.StatusNotFound)
			return
		}
		http.Error(res, fmt.Sprintf("failed to get metric %s / %s: %v", metricType, metricName, err),
			http.StatusBadRequest)
		return
	}

	var r string
	if m.MType == model.Gauge {
		r = strconv.FormatFloat(*m.Value, 'f', -1, 64)
	} else {
		r = fmt.Sprintf("%d", *m.Delta)
	}

	res.Header().Set("Content-Type", "text/html")
	_, err = res.Write([]byte(r))
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to write body: %v", err),
			http.StatusInternalServerError)
		return
	}
}

func (a *APIServer) List(res http.ResponseWriter, _ *http.Request) {
	var b bytes.Buffer
	metrics := a.storage.List()
	for id, m := range metrics {
		b.WriteString(id)
		switch m.MType {
		case model.Gauge:
			b.WriteString(fmt.Sprintf(" %.5f\n", *m.Value))
		case model.Counter:
			b.WriteString(fmt.Sprintf(" %d\n", *m.Delta))
		}
	}

	res.Header().Set("Content-Type", "text/html")
	_, err := res.Write(b.Bytes())
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to write body: %v", err),
			http.StatusInternalServerError)
		return
	}
}
