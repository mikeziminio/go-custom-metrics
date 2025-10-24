package memstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

func TestSync(t *testing.T) {
	testCases := []struct {
		name          string
		metrics       map[string]model.Metric
		expectedError bool
	}{
		{
			name: "синхронизация пустого хранилища",
			metrics: map[string]model.Metric{
				"test": {
					ID:    "test",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 1.5),
				},
			},
			expectedError: false,
		},
		{
			name: "синхронизация с несколькими метриками",
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
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем временный файл для тестирования
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "metrics.json")

			ms := New(false, tmpFile, zap.L())
			// Initialize with test metrics
			for _, metric := range tc.metrics {
				_, err := ms.Update(metric)
				require.NoError(t, err)
			}

			// Тестируем метод Sync
			err := ms.Sync()
			if tc.expectedError {
				require.Error(t, err)
			} else {
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
			}
		})
	}
}

func TestRestore(t *testing.T) {
	testCases := []struct {
		name            string
		setupFile       func(*testing.T, string)
		expectedError   bool
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
			expectedError: false,
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
		{
			name: "восстановление из несуществующего файла",
			setupFile: func(t *testing.T, _ string) {
				t.Helper()
				// Ничего не делаем, файл не будет существовать
			},
			// исходя из того как сделаны тесты от yandex - в этом случае
			// сервер просто должен стартовать с пустыми метриками
			expectedError:   false,
			expectedMetrics: make(map[string]model.Metric),
		},
		{
			name: "восстановление из файла с некорректным JSON",
			setupFile: func(t *testing.T, filename string) {
				t.Helper()
				err := os.WriteFile(filename, []byte("invalid json"), 0600)
				require.NoError(t, err)
			},
			expectedError:   true,
			expectedMetrics: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем временный файл для тестирования
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "metrics.json")

			// Подготавливаем тестовый файл
			tc.setupFile(t, tmpFile)

			ms := New(false, tmpFile, zap.L())

			// Тестируем метод Restore
			err := ms.Restore()
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Проверяем, что метрики были успешно восстановлены
				assert.Equal(t, tc.expectedMetrics, ms.List())
			}
		})
	}
}
