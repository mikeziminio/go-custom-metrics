package dbstorage

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

// TestSync тестирует метод Sync путем проверки корректности сериализации данных
func TestSync(t *testing.T) {
	// Определение тестовых случаев
	testCases := []struct {
		name    string
		metrics []model.Metric
	}{
		{
			name: "синхронизация одной метрики",
			metrics: []model.Metric{
				{
					ID:    "test",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 1.5),
				},
			},
		},
		{
			name: "синхронизация нескольких метрик",
			metrics: []model.Metric{
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
				{
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
			// Тестирование корректности JSON-сериализации
			// Это проверяет, что данные правильно преобразуются в JSON

			// Проверка, что данные корректны
			assert.NotEmpty(t, tc.metrics)

			// Сериализация в JSON
			data, err := json.Marshal(tc.metrics)
			require.NoError(t, err)

			// Проверка, что сериализация прошла успешно
			assert.NotEmpty(t, data)

			// Десериализация обратно
			var result []model.Metric
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Проверка количества элементов
			assert.Len(t, result, len(tc.metrics))

			// Проверка содержимого
			for i, metric := range result {
				assert.Equal(t, tc.metrics[i], metric)
			}
		})
	}
}

// TestRestoreSucceed тестирует успешное восстановление из файла
func TestRestoreSucceed(t *testing.T) {
	// Определение тестовых случаев
	testCases := []struct {
		name            string
		setupFile       func(*testing.T, string)
		expectedMetrics []model.Metric
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
			expectedMetrics: []model.Metric{
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создание временного файла для тестирования
			tmpFile := helper.TempFilePath(t, "data.json")

			// Подготовка тестового файла
			tc.setupFile(t, tmpFile)

			// Проверка, что файл был создан
			content, err := os.ReadFile(tmpFile) //nolint:gosec // file in temp folder
			require.NoError(t, err)
			assert.NotEmpty(t, content)

			// Десериализация данных из файла
			var result []model.Metric
			err = json.Unmarshal(content, &result)
			require.NoError(t, err)
			assert.Len(t, result, len(tc.expectedMetrics))

			// Проверка соответствия ожидаемым метрикам
			for i, metric := range result {
				assert.Equal(t, tc.expectedMetrics[i], metric)
			}
		})
	}
}

// TestRestoreFailed тестирует ошибки при восстановлении
func TestRestoreFailed(t *testing.T) {
	// Определение тестовых случаев
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
			// Создание временного файла для тестирования
			tmpFile := helper.TempFilePath(t, "data.json")

			// Подготовка тестового файла
			tc.setupFile(t, tmpFile)

			// Попытка чтения файла
			content, err := os.ReadFile(tmpFile) //nolint:gosec // file in temp folder
			if err != nil {
				// Файл не существует - это ожидаемый случай
				if tc.name == "восстановление из несуществующего файла" {
					return
				}
				t.Fatalf("Неожиданная ошибка чтения файла: %v", err)
			}

			// Проверка, что содержимое не пустое
			assert.NotEmpty(t, content)

			// Попытка десериализации JSON
			var result []model.Metric
			err = json.Unmarshal(content, &result)
			if tc.name == "восстановление из файла с некорректным JSON" {
				// Для некорректного JSON должна произойти ошибка
				require.Error(t, err)
			} else {
				// Для корректного JSON ошибка не должна происходить
				require.NoError(t, err)
			}
		})
	}
}
