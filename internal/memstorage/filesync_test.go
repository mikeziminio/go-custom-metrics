package memstorage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

func TestSync(t *testing.T) {
	testCases := []struct {
		name    string
		metrics map[string]model.Metric
	}{
		{
			name: "синхронизация одной метрики",
			metrics: map[string]model.Metric{
				"test": {
					ID:    "test",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 1.5),
				},
			},
		},
		{
			name: "синхронизация нескольких метрик",
			metrics: map[string]model.Metric{
				"counter1": {
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 10),
					Value: nil,
				},
				"gauge1": {
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 2.5),
				},
				"gauge2": {
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 3.5),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем временный файл для тестирования
			tmpFile := helper.TempFilePath(t, "data.json")

			ms, err := New(false, false, tmpFile, zap.L())
			require.NoError(t, err)
			// Initialize with test metrics
			for _, metric := range tc.metrics {
				_, err := ms.Update(metric)
				require.NoError(t, err)
			}

			// Тестируем метод Sync
			err = ms.Sync()
			require.NoError(t, err)

			// Проверяем, что файл был создан и содержит ожидаемые данные
			content, err := os.ReadFile(tmpFile) //nolint:gosec // file in temp folder
			require.NoError(t, err)
			require.NotEmpty(t, content)

			// Разбираем содержимое для проверки, что это корректный JSON
			var result []model.Metric
			err = json.Unmarshal(content, &result)
			require.NoError(t, err)
			require.Len(t, result, len(tc.metrics))

			// Проверяем, что метрики в файле соответствуют оригинальным
			for _, metric := range result {
				expectedMetric, exists := tc.metrics[metric.ID]
				require.True(t, exists)
				assert.Equal(t, expectedMetric, metric)
			}
		})
	}
}

func TestRestoreSucceed(t *testing.T) {
	testCases := []struct {
		name            string
		setupFile       func(*testing.T, string)
		expectedMetrics map[string]model.Metric
	}{
		{
			name: "восстановление из корректного файла",
			setupFile: func(t *testing.T, filename string) {
				t.Helper()
				metrics := []model.Metric{
					{
						ID:    "counter1",
						MType: model.Counter,
						Delta: helper.NewInt64(t, 10),
						Value: nil,
					},
					{
						ID:    "gauge1",
						MType: model.Gauge,
						Delta: nil,
						Value: helper.NewFloat64(t, 2.5),
					},
				}
				content, err := json.Marshal(metrics)
				require.NoError(t, err)
				err = os.WriteFile(filename, content, 0600)
				require.NoError(t, err)
			},
			expectedMetrics: map[string]model.Metric{
				"counter1": {
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 10),
					Value: nil,
				},
				"gauge1": {
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 2.5),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем временный файл для тестирования
			tmpFile := helper.TempFilePath(t, "data.json")

			// Подготавливаем тестовый файл
			tc.setupFile(t, tmpFile)

			// Тестируем механизм restore
			ms, err := New(false, true, tmpFile, zap.L())
			require.NoError(t, err)
			// Проверяем, что метрики были успешно восстановлены
			assert.Equal(t, tc.expectedMetrics, ms.List())
		})
	}
}

func TestRestoreFailed(t *testing.T) {
	testCases := []struct {
		name      string
		setupFile func(*testing.T, string)
	}{
		{
			name: "восстановление из несуществующего файла",
			setupFile: func(t *testing.T, _ string) {
				t.Helper()
				// Ничего не делаем, файл не будет существовать
			},
		},
		{
			name: "восстановление из файла с некорректным JSON",
			setupFile: func(t *testing.T, filename string) {
				t.Helper()
				err := os.WriteFile(filename, []byte("invalid json"), 0600)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем временный файл для тестирования
			tmpFile := helper.TempFilePath(t, "data.json")

			// Подготавливаем тестовый файл
			tc.setupFile(t, tmpFile)

			// Тестируем механизм restore
			s := MemStorage{
				logger: zap.L(),
			}
			err := s.restore()
			require.Error(t, err)
		})
	}
}
