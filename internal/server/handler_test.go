package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name               string
		metricType         model.MetricType
		metricName         string
		metricValue        string
		metric             *model.Metric
		storageReturnError error
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:        "update counter value",
			metricType:  model.Counter,
			metricName:  "some",
			metricValue: "8",
			metric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			storageReturnError: nil,
			expectedStatus:     200,
		},
		{
			name:        "update gauge value",
			metricType:  model.Gauge,
			metricName:  "some",
			metricValue: "8.1234",
			metric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 8.1234),
			},
			storageReturnError: nil,
			expectedStatus:     200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Update(*tc.metric).
				Return(tc.storageReturnError).
				Once()

			server := New("", storage, zap.L())
			server.RegisterRoutes()

			path := fmt.Sprintf("/update/%s/%s/%s", tc.metricType, tc.metricName, tc.metricValue)
			req := httptest.NewRequest(http.MethodPost, path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name                string
		metricType          model.MetricType
		metricName          string
		storageReturnMetric *model.Metric
		storageReturnError  error
		expectedStatus      int
		expectedBody        string
	}{
		{
			name:                "counter value not found",
			metricType:          model.Counter,
			metricName:          "some",
			storageReturnMetric: nil,
			storageReturnError:  model.ErrMetricNotFound,
			expectedStatus:      404,
			expectedBody:        "",
		},
		{
			name:       "counter value",
			metricType: model.Counter,
			metricName: "some",
			storageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			storageReturnError: nil,
			expectedStatus:     200,
			expectedBody:       "8",
		},
		{
			name:       "gauge value",
			metricType: model.Gauge,
			metricName: "some",
			storageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 64.555),
			},
			storageReturnError: nil,
			expectedStatus:     200,
			expectedBody:       "64.555",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Get(tc.metricType, tc.metricName).
				Return(tc.storageReturnMetric, tc.storageReturnError).
				Once()

			server := New("", storage, zap.L())
			server.RegisterRoutes()

			path := fmt.Sprintf("/value/%s/%s", tc.metricType, tc.metricName)
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			body, _ := io.ReadAll(rec.Body)
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}

func TestList(t *testing.T) {
	storage := NewMockStorage(t)
	storage.EXPECT().List().
		Return(map[string]model.Metric{
			"some": {
				ID:    "some",
				MType: model.Gauge,
				Value: helper.NewFloat64(t, 8.12345),
			},
			"other": {
				ID:    "other",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 64),
			},
		}).
		Once()

	server := New("", storage, zap.L())
	server.RegisterRoutes()

	path := "/"
	req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "some 8.12345")
	assert.Contains(t, string(body), "other 64")
}
