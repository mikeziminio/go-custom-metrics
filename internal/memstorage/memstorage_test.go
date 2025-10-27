package memstorage

import (
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
			ms, err := New(false, false, "", zap.L())
			require.NoError(t, err)
			// Initialize with initial metrics
			for _, v := range tc.initialMetrics {
				_, err := ms.Update(v)
				require.NoError(t, err)
			}

			_, err = ms.Update(tc.updatedModel)
			require.NoError(t, err)

			// Test by retrieving all metrics and comparing
			result := ms.List()
			assert.Equal(t, tc.expectedMetrics, result)
		})
	}
}

func TestGetCounter(t *testing.T) {
	ms, err := New(false, false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 1),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "other",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 2),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 3),
		Value: nil,
	})
	require.NoError(t, err)
	m, err := ms.Get(model.Counter, "some")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 4),
		Value: nil,
	}, m)
	m, err = ms.Get(model.Counter, "other")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "other",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 2),
		Value: nil,
	}, m)
}

func TestGetGauge(t *testing.T) {
	ms, err := New(false, false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 1),
	})
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 2),
	})
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 3),
	})
	require.NoError(t, err)
	m, err := ms.Get(model.Gauge, "some")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "some",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 3),
	}, m)
	m, err = ms.Get(model.Gauge, "other")
	require.NoError(t, err)
	assert.Equal(t, &model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 2),
	}, m)
}

func TestList(t *testing.T) {
	ms, err := New(false, false, "", zap.L())
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "some",
		MType: model.Counter,
		Delta: helper.NewInt64(t, 1),
		Value: nil,
	})
	require.NoError(t, err)
	_, err = ms.Update(model.Metric{
		ID:    "other",
		MType: model.Gauge,
		Delta: nil,
		Value: helper.NewFloat64(t, 88),
	})
	require.NoError(t, err)

	// Test List method - it should not return an error
	m := ms.List()
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
