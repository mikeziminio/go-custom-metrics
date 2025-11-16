package memstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name            string
		initialMetrics  map[string]model.Metric
		updatedModel    model.Metric
		expectedMetrics map[string]model.Metric
	}{
		{
			name:           "add counter metric to empty map",
			initialMetrics: make(map[string]model.Metric),
			updatedModel: model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 5),
				Value: nil,
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
		},
		{
			name: "add counter metric",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedModel: model.Metric{
				ID:    "other",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				"other": {
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
			},
		},
		{
			name:           "add gauge metric to empty map",
			initialMetrics: make(map[string]model.Metric),
			updatedModel: model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 5),
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5),
				},
			},
		},
		{
			name: "add gauge metric",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5),
				},
			},
			updatedModel: model.Metric{
				ID:    "other",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 8),
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5),
				},
				"other": {
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
			},
		},
		{
			name: "update gauge metric",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
			},
			updatedModel: model.Metric{
				ID:    "some",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 8),
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
			},
		},
		{
			name: "update counter metric",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedModel: model.Metric{
				ID:    "some",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 13),
					Value: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ms, err := New(false, "", zap.L())
			require.NoError(t, err)
			// Initialize with initial metrics
			for _, v := range tc.initialMetrics {
				_, err := ms.Update(context.Background(), v)
				require.NoError(t, err)
			}

			_, err = ms.Update(context.Background(), tc.updatedModel)
			require.NoError(t, err)

			// Test by retrieving all metrics and comparing
			result, err := ms.List(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMetrics, result)
		})
	}
}

func TestGetCounter(t *testing.T) {
	ms, err := New(false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 1),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "other",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 2),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 3),
		Value: nil,
	})
	require.NoError(t, err)
	m, err := ms.Get(context.Background(), model.Counter, "some")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 4),
		Value: nil,
	}, m)
	m, err = ms.Get(context.Background(), model.Counter, "other")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "other",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 2),
		Value: nil,
	}, m)
}

func TestGetGauge(t *testing.T) {
	ms, err := New(false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 1),
	})
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 2),
	})
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 3),
	})
	require.NoError(t, err)
	m, err := ms.Get(context.Background(), model.Gauge, "some")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 3),
	}, m)
	m, err = ms.Get(context.Background(), model.Gauge, "other")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 2),
	}, m)
}

func TestList(t *testing.T) {
	ms, err := New(false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 1),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(context.Background(), model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 88),
	})
	require.NoError(t, err)

	// Test List method - it should not return an error
	m, err := ms.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, map[string]model.Metric{
		"some": {
			ID:    "some",
			MType: model.Counter,
			Delta: helper.NewInt64(t, 1),
			Value: nil,
		},
		"other": {
			ID:    "other",
			MType: model.Gauge,
			Delta: nil,
			Value: helper.NewFloat64(t, 88),
		},
	}, m)
}

func TestUpdates(t *testing.T) {
	testCases := []struct {
		name            string
		initialMetrics  map[string]model.Metric
		updatedMetrics  []model.Metric
		expectedMetrics map[string]model.Metric
	}{
		{
			name:           "add counter metrics to empty map",
			initialMetrics: make(map[string]model.Metric),
			updatedMetrics: []model.Metric{
				{
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				{
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				"other": {
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
			},
		},
		{
			name: "add gauge metrics",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
				{
					ID:    "third",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 12),
				},
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				"other": {
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
				"third": {
					ID:    "third",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 12),
				},
			},
		},
		{
			name: "update existing counter metrics",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				"other": {
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 3),
					Value: nil,
				},
				{
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 2),
					Value: nil,
				},
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
				"other": {
					ID:    "other",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 10),
					Value: nil,
				},
			},
		},
		{
			name: "update existing gauge metrics",
			initialMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5),
				},
				"other": {
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8),
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 15),
				},
				{
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 20),
				},
			},
			expectedMetrics: map[string]model.Metric{
				"some": {
					ID:    "some",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 15),
				},
				"other": {
					ID:    "other",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 20),
				},
			},
		},
		{
			name: "mixed updates",
			initialMetrics: map[string]model.Metric{
				"counter1": {
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				"gauge1": {
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 10),
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 3),
					Value: nil,
				},
				{
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 15),
				},
				{
					ID:    "counter2",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 7),
					Value: nil,
				},
				{
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 25),
				},
			},
			expectedMetrics: map[string]model.Metric{
				"counter1": {
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
				"gauge1": {
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 15),
				},
				"counter2": {
					ID:    "counter2",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 7),
					Value: nil,
				},
				"gauge2": {
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 25),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ms, err := New(false, "", zap.L())
			require.NoError(t, err)
			// Initialize with initial metrics
			for _, v := range tc.initialMetrics {
				_, err := ms.Update(context.Background(), v)
				require.NoError(t, err)
			}

			// Call Updates method
			err = ms.Updates(context.Background(), tc.updatedMetrics)
			require.NoError(t, err)

			// Test by retrieving all metrics and comparing
			result, err := ms.List(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMetrics, result)
		})
	}
}
