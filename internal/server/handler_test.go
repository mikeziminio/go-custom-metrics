package server

import (
	"bytes"
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
		name             string
		body             string
		expectedMetric   *model.Metric
		expectedStatus   int
		expectedJSONBody string
	}{
		{
			name: "update counter value",
			body: `{
				"id": "some",
				"type": "counter",
				"delta": 8
			}`,
			expectedMetric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
			},
			expectedStatus: http.StatusOK,
			expectedJSONBody: `{
				"id": "some",
				"type": "counter",
				"delta": 8
			}`,
		},
		{
			name: "update gauge value",
			body: `{
				"id":"some",
				"type":"gauge",
				"value":8.1234
			}`,
			expectedMetric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Value: helper.NewFloat64(t, 8.1234),
			},
			expectedStatus: http.StatusOK,
			expectedJSONBody: `{
				"id":"some",
				"type":"gauge",
				"value":8.1234
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Update(*tc.expectedMetric).
				Return(tc.expectedMetric, nil).
				Once()

			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			path := "/update"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			body, _ := io.ReadAll(rec.Body)
			assert.JSONEq(t, tc.expectedJSONBody, string(body))
		})
	}
}

func TestUpdateFailed(t *testing.T) {
	testCases := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name: "update with incorrect data",
			body: `{
				"id": "some",
				"type": "gauge",
				"delta": "incorrect value"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update with invalid json",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			path := "/update"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestUpdateByParams(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expectedMetric *model.Metric
		expectedStatus int
	}{
		{
			name: "update counter value",
			path: "/update/counter/some/8",
			expectedMetric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
			},
		},
		{
			name: "update gauge value",
			path: "/update/gauge/some/8.1234",
			expectedMetric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Value: helper.NewFloat64(t, 8.1234),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Update(*tc.expectedMetric).
				Return(tc.expectedMetric, nil).
				Once()

			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			req := httptest.NewRequest(http.MethodPost, tc.path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestUpdateByParamsFailed(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "update with incorrect data",
			path:           "/update/counter/some/8.677",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update with incorrect path",
			path:           "/update/incorrect_path",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			req := httptest.NewRequest(http.MethodPost, tc.path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name                    string
		body                    string
		expectMetricType        model.MetricType
		expectMetricName        string
		mockStorageReturnMetric *model.Metric
		expectJSONBody          string
	}{
		{
			name: "counter value",
			body: `{
				"id": "some",
				"type": "counter"
			}`,
			expectMetricType: model.Counter,
			expectMetricName: "some",
			mockStorageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectJSONBody: `{"id":"some","type":"counter","delta":8}`,
		},
		{
			name: "gauge value",
			body: `{
				"id": "some",
				"type": "gauge"
			}`,
			expectMetricType: model.Gauge,
			expectMetricName: "some",
			mockStorageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 64.555),
			},
			expectJSONBody: `{"id":"some","type":"gauge","value":64.555}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Get(tc.expectMetricType, tc.expectMetricName).
				Return(tc.mockStorageReturnMetric, nil).
				Once()

			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			path := "/value"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			body, _ := io.ReadAll(rec.Body)
			assert.JSONEq(t, tc.expectJSONBody, string(body))
		})
	}
}

func TestGetFailed(t *testing.T) {
	testCases := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "update with invalid json",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			path := "/value"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestGetByParams(t *testing.T) {
	testCases := []struct {
		name                    string
		path                    string
		expectMetricType        model.MetricType
		expectMetricName        string
		mockStorageReturnMetric *model.Metric
		expectTextBody          string
	}{
		{
			name:             "counter value",
			path:             "/value/counter/some",
			expectMetricType: model.Counter,
			expectMetricName: "some",
			mockStorageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectTextBody: "8",
		},
		{
			name:             "gauge value",
			path:             "/value/gauge/some",
			expectMetricType: model.Gauge,
			expectMetricName: "some",
			mockStorageReturnMetric: &model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 64.555),
			},
			expectTextBody: "64.555",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			storage.EXPECT().Get(tc.expectMetricType, tc.expectMetricName).
				Return(tc.mockStorageReturnMetric, nil).
				Once()

			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			req := httptest.NewRequest(http.MethodGet, tc.path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			body, _ := io.ReadAll(rec.Body)
			assert.Equal(t, tc.expectTextBody, string(body))
		})
	}
}

func TestGetByParamsFailed(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "get with incorrect path",
			path:           "/value/incorrect_path",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockStorage(t)
			server := New("", 0, storage, zap.L())
			server.RegisterRoutes()

			req := httptest.NewRequest(http.MethodPost, tc.path, http.NoBody)
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedStatus, rec.Code)
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

	server := New("", 0, storage, zap.L())
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
