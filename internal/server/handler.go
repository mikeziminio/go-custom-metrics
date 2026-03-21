package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

// Update handles HTTP POST /value request to update a single metric.
// It reads the metric from request body, updates it in storage,
// and logs an audit event if enabled.
func (a *APIServer) Update(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	var data updateReqSchema

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to validate request body: %v", err),
			http.StatusBadRequest)
		return
	}

	m, err := a.storage.Update(req.Context(), model.Metric{
		ID:    data.ID,
		MType: data.MType,
		Delta: data.Delta,
		Value: data.Value,
	})
	if err != nil {
		if errors.Is(err, model.ErrIncorrectMetricType) {
			http.Error(res, fmt.Sprintf("failed to fetch metric type: %v", data.MType),
				http.StatusBadRequest,
			)
			return
		}
		a.handleInternalServerError(res,
			fmt.Errorf("failed to update metric value %s / %s: %w", data.MType, data.ID, err),
		)
		return
	}

	// Log audit event
	if a.auditLogger != nil {
		ipAddress := extractIPAddress(req)
		event := AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   []string{fmt.Sprintf("%s:%s", data.MType, data.ID)},
			IPAddress: ipAddress,
		}
		if err := a.auditLogger.Log(req.Context(), event); err != nil {
			a.logger.Error("Failed to log audit event", zap.Error(err))
		}
	}

	resData, err := json.Marshal(m)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(resData)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to write response: %w", err))
		return
	}
}

// Updates handles HTTP POST /updates request to update multiple metrics.
// It reads metrics from request body and updates them in storage in bulk,
// then logs an audit event if enabled.
func (a *APIServer) Updates(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	var data updatesReqSchema
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to validate request body: %v", err),
			http.StatusBadRequest)
		return
	}

	metrics := make([]model.Metric, 0, len(data))
	for _, d := range data {
		metrics = append(metrics, model.Metric{
			ID:    d.ID,
			MType: d.MType,
			Delta: d.Delta,
			Value: d.Value,
		})
	}

	err = a.storage.Updates(req.Context(), metrics)

	if err != nil {
		if errors.Is(err, model.ErrIncorrectMetricType) {
			http.Error(res, "failed to fetch metric type",
				http.StatusBadRequest,
			)
			return
		}
		a.handleInternalServerError(res,
			fmt.Errorf("failed to update metric values: %w", err),
		)
		return
	}

	// Log audit event
	if a.auditLogger != nil {
		ipAddress := extractIPAddress(req)
		metricNames := make([]string, 0, len(data))
		for _, d := range data {
			metricNames = append(metricNames, fmt.Sprintf("%s:%s", d.MType, d.ID))
		}
		event := AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: ipAddress,
		}
		if err := a.auditLogger.Log(req.Context(), event); err != nil {
			a.logger.Error("Failed to log audit event", zap.Error(err))
		}
	}

	res.WriteHeader(http.StatusOK)
}

// UpdateByParams handles HTTP POST /update/{metricType}/{metricName}/{value} request.
// It extracts metric parameters from URL and updates the metric in storage,
// then logs an audit event if enabled.
func (a *APIServer) UpdateByParams(res http.ResponseWriter, req *http.Request) {
	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to fetch metric type: %v", err),
			http.StatusBadRequest,
		)
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
				http.StatusBadRequest,
			)
			return
		}
		delta = &d
	}

	if metricType == model.Gauge {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("failed to parse gauge value: %v", err),
				http.StatusBadRequest,
			)
			return
		}
		value = &v
	}

	_, err = a.storage.Update(req.Context(), model.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: delta,
		Value: value,
	})
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to update metric value: %w", err))
		return
	}

	// Log audit event
	if a.auditLogger != nil {
		ipAddress := extractIPAddress(req)
		event := AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   []string{fmt.Sprintf("%s:%s", metricType, metricName)},
			IPAddress: ipAddress,
		}
		if err := a.auditLogger.Log(req.Context(), event); err != nil {
			a.logger.Error("Failed to log audit event", zap.Error(err))
		}
	}

	res.WriteHeader(http.StatusOK)
}

// Get handles HTTP POST /value request to retrieve a single metric.
// It reads metric type and name from request body and returns the value.
func (a *APIServer) Get(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	var data getReqSchema

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to validate request body: %v", err),
			http.StatusBadRequest)
		return
	}

	m, err := a.storage.Get(req.Context(), data.MType, data.ID)
	if err != nil {
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(res, fmt.Sprintf("metric not found %s / %s: %v", data.MType, data.ID, err),
				http.StatusNotFound,
			)
			return
		}
		a.handleInternalServerError(res, fmt.Errorf("failed to get metric: %w", err))
		return
	}

	resData, err := json.Marshal(m)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(resData)
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to write response: %w", err))
		return
	}
}

// GetByParams handles HTTP GET /value/{metricType}/{metricName} request.
// It extracts metric parameters from URL and returns the value as plain text.
func (a *APIServer) GetByParams(res http.ResponseWriter, req *http.Request) {
	mt := chi.URLParam(req, "metricType")
	metricType, err := model.NewMetricTypeFromString(mt)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to fetch metric type: %v", err),
			http.StatusBadRequest,
		)
		return
	}
	metricName := chi.URLParam(req, "metricName")

	m, err := a.storage.Get(req.Context(), metricType, metricName)
	if err != nil {
		if errors.Is(err, model.ErrMetricNotFound) {
			http.Error(res, fmt.Sprintf("metric not found %s / %s: %v", metricType, metricName, err),
				http.StatusNotFound,
			)
			return
		}
		a.handleInternalServerError(res, fmt.Errorf("failed to get metric %s / %s: %w", metricType, metricName, err))
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
		a.handleInternalServerError(res, fmt.Errorf("failed to write response: %w", err))
		return
	}
}

// List handles HTTP GET / request to list all metrics.
// It returns metrics in HTML format with id and value.
func (a *APIServer) List(res http.ResponseWriter, req *http.Request) {
	var b bytes.Buffer
	metrics, err := a.storage.List(req.Context())
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to get metrics list: %w", err))
		return
	}
	for id, m := range metrics {
		b.WriteString(id)
		b.WriteString(" ")
		switch m.MType {
		case model.Gauge:
			b.WriteString(strconv.FormatFloat(*m.Value, 'f', 5, 64))
		case model.Counter:
			b.WriteString(strconv.FormatInt(*m.Delta, 10))
		}
	}

	res.Header().Set("Content-Type", "text/html")
	_, err = res.Write(b.Bytes())
	if err != nil {
		a.handleInternalServerError(res, fmt.Errorf("failed to write response: %w", err))
		return
	}
}

// Ping handles HTTP GET /ping health check request.
// It returns status 200 if storage is reachable.
func (a *APIServer) Ping(res http.ResponseWriter, req *http.Request) {
	err := a.storage.Ping(req.Context())
	if err != nil {
		a.handleInternalServerError(res, err)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (a *APIServer) handleInternalServerError(res http.ResponseWriter, err error) {
	a.logger.Error("Internal server error", zap.Error(err))
	status := http.StatusInternalServerError
	http.Error(res, http.StatusText(status), status)
}
