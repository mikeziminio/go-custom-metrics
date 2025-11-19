package dbstorage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/test/helper"
)

func testPostgresContainer(t *testing.T, ctx context.Context) *postgres.PostgresContainer {
	pc, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	require.NoError(t, err)
	return pc
}

func testDBStorage(t *testing.T, ctx context.Context) (*DBStorage, *postgres.PostgresContainer, string) {
	// Start PostgreSQL container
	pc := testPostgresContainer(t, ctx)

	// Get connection string
	connStr, err := pc.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create DBStorage instance
	logger, _ := zap.NewDevelopment()
	dbStorage, err := New(connStr, false, "", logger)
	require.NoError(t, err)

	return dbStorage, pc, connStr
}

func TestDBStorage_MigrateUp(t *testing.T) {
	ctx := context.Background()

	// Run migrations
	dbStorage, pc, connString := testDBStorage(t, ctx)
	defer dbStorage.Close()
	defer pc.Terminate(ctx)

	// Verify that the table exists
	var count int
	err := dbStorage.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM metric").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count) // Table should exist but be empty

	// Try to run migration again (should not error)
	err = dbStorage.migrateUp(connString)
	require.NoError(t, err)
}

func TestDBStorage_Update(t *testing.T) {
	ctx := context.Background()

	dbStorage, pc, _ := testDBStorage(t, ctx)
	defer dbStorage.Close()
	defer pc.Terminate(ctx)

	// Test cases
	testCases := []struct {
		name           string
		initialMetrics []model.Metric
		updatedMetric  model.Metric
		expectedError  bool
		expectedDelta  *int64
		expectedValue  *float64
	}{
		{
			name: "adding counter to empty database",
			updatedMetric: model.Metric{
				ID:    "some_counter",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 5),
				Value: nil,
			},
			expectedDelta: helper.NewInt64(t, 5),
			expectedValue: nil,
		},
		{
			name: "adding counter",
			initialMetrics: []model.Metric{
				{
					ID:    "existing_counter",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedMetric: model.Metric{
				ID:    "new_counter",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectedDelta: helper.NewInt64(t, 8),
			expectedValue: nil,
		},
		{
			name: "adding gauge to empty database",
			updatedMetric: model.Metric{
				ID:    "some_gauge",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 5.0),
			},
			expectedDelta: nil,
			expectedValue: helper.NewFloat64(t, 5.0),
		},
		{
			name: "adding gauge",
			initialMetrics: []model.Metric{
				{
					ID:    "existing_gauge",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5.0),
				},
			},
			updatedMetric: model.Metric{
				ID:    "new_gauge",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 8.0),
			},
			expectedDelta: nil,
			expectedValue: helper.NewFloat64(t, 8.0),
		},
		{
			name: "updating gauge",
			initialMetrics: []model.Metric{
				{
					ID:    "update_gauge",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8.0),
				},
			},
			updatedMetric: model.Metric{
				ID:    "update_gauge",
				MType: model.Gauge,
				Delta: nil,
				Value: helper.NewFloat64(t, 15.0),
			},
			expectedDelta: nil,
			expectedValue: helper.NewFloat64(t, 15.0),
		},
		{
			name: "updating counter",
			initialMetrics: []model.Metric{
				{
					ID:    "update_counter",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedMetric: model.Metric{
				ID:    "update_counter",
				MType: model.Counter,
				Delta: helper.NewInt64(t, 8),
				Value: nil,
			},
			expectedDelta: helper.NewInt64(t, 13),
			expectedValue: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Чистим таблицу для каждого теста
			_, err := dbStorage.db.ExecContext(ctx, "DELETE FROM metric")
			require.NoError(t, err)

			// Добавляем исходные метрики из тест-кейса из initialMetrics
			for _, metric := range tc.initialMetrics {
				_, err = dbStorage.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, delta, value)
					VALUES ($1, $2, $3, $4)`,
					metric.ID, metric.MType, metric.Delta, metric.Value)
				require.NoError(t, err)
			}

			// Обновляем одну метрику из updatedMetric
			result, err := dbStorage.Update(ctx, tc.updatedMetric)
			if tc.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.updatedMetric.ID, result.ID)
			assert.Equal(t, tc.updatedMetric.MType, result.MType)

			// Проверяем обновленные метрики
			var storedID string
			var storedType string
			var storedDelta sql.NullInt64
			var storedValue sql.NullFloat64

			err = dbStorage.db.QueryRowContext(ctx, `
				SELECT id, m_type, delta, value
				FROM metric WHERE id = $1 AND m_type = $2`,
				tc.updatedMetric.ID, tc.updatedMetric.MType,
			).
				Scan(&storedID, &storedType, &storedDelta, &storedValue)
			require.NoError(t, err)

			assert.Equal(t, tc.updatedMetric.ID, storedID)
			assert.Equal(t, tc.updatedMetric.MType, model.MetricType(storedType))

			// Check delta value
			if tc.expectedDelta != nil {
				assert.True(t, storedDelta.Valid)
				assert.Equal(t, *tc.expectedDelta, storedDelta.Int64)
			} else {
				assert.False(t, storedDelta.Valid)
			}

			// Check value
			if tc.expectedValue != nil {
				assert.True(t, storedValue.Valid)
				assert.Equal(t, *tc.expectedValue, storedValue.Float64)
			} else {
				assert.False(t, storedValue.Valid)
			}
		})
	}
}

func TestDBStorage_Get(t *testing.T) {
	ctx := context.Background()

	dbStorage, pc, _ := testDBStorage(t, ctx)
	defer dbStorage.Close()
	defer pc.Terminate(ctx)

	// Test cases
	testCases := []struct {
		name          string
		setupMetrics  []model.Metric
		metricType    model.MetricType
		metricName    string
		expectFound   bool
		expectedDelta *int64
		expectedValue *float64
	}{
		{
			name:        "getting non-existent metric",
			metricType:  model.Counter,
			metricName:  "non_existent",
			expectFound: false,
		},
		{
			name: "getting counter",
			setupMetrics: []model.Metric{
				{
					ID:    "test_counter",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 100),
					Value: nil,
				},
			},
			metricType:    model.Counter,
			metricName:    "test_counter",
			expectFound:   true,
			expectedDelta: helper.NewInt64(t, 100),
			expectedValue: nil,
		},
		{
			name: "getting gauge",
			setupMetrics: []model.Metric{
				{
					ID:    "test_gauge",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 99.5),
				},
			},
			metricType:    model.Gauge,
			metricName:    "test_gauge",
			expectFound:   true,
			expectedDelta: nil,
			expectedValue: helper.NewFloat64(t, 99.5),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear database for each test
			_, err := dbStorage.db.ExecContext(ctx, "DELETE FROM metric")
			require.NoError(t, err)

			// Insert test metrics if any
			for _, metric := range tc.setupMetrics {
				_, err = dbStorage.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, delta, value)
					VALUES ($1, $2, $3, $4)`,
					metric.ID, metric.MType, metric.Delta, metric.Value)
				require.NoError(t, err)
			}

			// Call Get method
			result, err := dbStorage.Get(ctx, tc.metricType, tc.metricName)

			if !tc.expectFound {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.metricName, result.ID)
			assert.Equal(t, tc.metricType, result.MType)

			// Check delta
			if tc.expectedDelta != nil {
				assert.NotNil(t, result.Delta)
				assert.Equal(t, *tc.expectedDelta, *result.Delta)
			} else {
				assert.Nil(t, result.Delta)
			}

			// Check value
			if tc.expectedValue != nil {
				assert.NotNil(t, result.Value)
				assert.Equal(t, *tc.expectedValue, *result.Value)
			} else {
				assert.Nil(t, result.Value)
			}
		})
	}
}

func TestDBStorage_List(t *testing.T) {
	ctx := context.Background()

	dbStorage, pc, _ := testDBStorage(t, ctx)
	defer dbStorage.Close()
	defer pc.Terminate(ctx)

	// Test cases
	testCases := []struct {
		name         string
		setupMetrics []model.Metric
		expectedLen  int
	}{
		{
			name:        "listing empty database",
			expectedLen: 0,
		},
		{
			name: "listing with metrics",
			setupMetrics: []model.Metric{
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
					Value: helper.NewFloat64(t, 5.5),
				},
			},
			expectedLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear database for each test
			_, err := dbStorage.db.ExecContext(ctx, "DELETE FROM metric")
			require.NoError(t, err)

			// Insert test metrics if any
			for _, metric := range tc.setupMetrics {
				_, err = dbStorage.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, delta, value)
					VALUES ($1, $2, $3, $4)`,
					metric.ID, metric.MType, metric.Delta, metric.Value)
				require.NoError(t, err)
			}

			// Call List method
			result, err := dbStorage.List(ctx)
			require.NoError(t, err)
			assert.Len(t, result, tc.expectedLen)

			// Check if all metrics are correctly returned
			for _, expectedMetric := range tc.setupMetrics {
				actualMetric, exists := result[expectedMetric.ID]
				assert.True(t, exists)
				assert.Equal(t, expectedMetric.ID, actualMetric.ID)
				assert.Equal(t, expectedMetric.MType, actualMetric.MType)

				if expectedMetric.Delta != nil {
					assert.NotNil(t, actualMetric.Delta)
					assert.Equal(t, *expectedMetric.Delta, *actualMetric.Delta)
				} else {
					assert.Nil(t, actualMetric.Delta)
				}

				if expectedMetric.Value != nil {
					assert.NotNil(t, actualMetric.Value)
					assert.Equal(t, *expectedMetric.Value, *actualMetric.Value)
				} else {
					assert.Nil(t, actualMetric.Value)
				}
			}
		})
	}
}

func TestDBStorage_Updates(t *testing.T) {
	ctx := context.Background()

	dbStorage, pc, _ := testDBStorage(t, ctx)
	defer dbStorage.Close()
	defer pc.Terminate(ctx)

	// Test cases
	testCases := []struct {
		name           string
		initialMetrics []model.Metric
		updatedMetrics []model.Metric
		expectedError  bool
		expectedDeltas map[string]*int64
		expectedValues map[string]*float64
	}{
		{
			name: "adding counter metrics to empty database",
			updatedMetrics: []model.Metric{
				{
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				{
					ID:    "counter2",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
				},
			},
			expectedDeltas: map[string]*int64{
				"counter1": helper.NewInt64(t, 5),
				"counter2": helper.NewInt64(t, 8),
			},
			expectedValues: map[string]*float64{},
		},
		{
			name: "adding gauge metrics",
			initialMetrics: []model.Metric{
				{
					ID:    "existing_counter",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8.0),
				},
				{
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 12.0),
				},
			},
			expectedDeltas: map[string]*int64{
				"existing_counter": helper.NewInt64(t, 5),
			},
			expectedValues: map[string]*float64{
				"gauge1": helper.NewFloat64(t, 8.0),
				"gauge2": helper.NewFloat64(t, 12.0),
			},
		},
		{
			name: "updating existing counter metrics",
			initialMetrics: []model.Metric{
				{
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				{
					ID:    "counter2",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 8),
					Value: nil,
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
					ID:    "counter2",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 2),
					Value: nil,
				},
			},
			expectedDeltas: map[string]*int64{
				"counter1": helper.NewInt64(t, 8),
				"counter2": helper.NewInt64(t, 10),
			},
			expectedValues: map[string]*float64{},
		},
		{
			name: "updating existing gauge metrics",
			initialMetrics: []model.Metric{
				{
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 5.0),
				},
				{
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 8.0),
				},
			},
			updatedMetrics: []model.Metric{
				{
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 15.0),
				},
				{
					ID:    "gauge2",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 20.0),
				},
			},
			expectedDeltas: map[string]*int64{},
			expectedValues: map[string]*float64{
				"gauge1": helper.NewFloat64(t, 15.0),
				"gauge2": helper.NewFloat64(t, 20.0),
			},
		},
		{
			name: "mixed updates",
			initialMetrics: []model.Metric{
				{
					ID:    "counter1",
					MType: model.Counter,
					Delta: helper.NewInt64(t, 5),
					Value: nil,
				},
				{
					ID:    "gauge1",
					MType: model.Gauge,
					Delta: nil,
					Value: helper.NewFloat64(t, 10.0),
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
					Value: helper.NewFloat64(t, 15.0),
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
					Value: helper.NewFloat64(t, 25.0),
				},
			},
			expectedDeltas: map[string]*int64{
				"counter1": helper.NewInt64(t, 8),
				"counter2": helper.NewInt64(t, 7),
			},
			expectedValues: map[string]*float64{
				"gauge1": helper.NewFloat64(t, 15.0),
				"gauge2": helper.NewFloat64(t, 25.0),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear database for each test
			_, err := dbStorage.db.ExecContext(ctx, "DELETE FROM metric")
			require.NoError(t, err)

			// Insert initial metrics if any
			for _, metric := range tc.initialMetrics {
				_, err = dbStorage.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, delta, value)
					VALUES ($1, $2, $3, $4)`,
					metric.ID, metric.MType, metric.Delta, metric.Value)
				require.NoError(t, err)
			}

			// Call Updates method
			err = dbStorage.Updates(ctx, tc.updatedMetrics)
			if tc.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Validate stored metrics
			for metricID, expectedDelta := range tc.expectedDeltas {
				var storedDelta sql.NullInt64
				err = dbStorage.db.QueryRowContext(ctx, `
					SELECT delta FROM metric WHERE id = $1 AND m_type = 'counter'`, metricID).
					Scan(&storedDelta)
				require.NoError(t, err)
				assert.True(t, storedDelta.Valid)
				assert.Equal(t, *expectedDelta, storedDelta.Int64)
			}

			for metricID, expectedValue := range tc.expectedValues {
				var storedValue sql.NullFloat64
				err = dbStorage.db.QueryRowContext(ctx, `
					SELECT value FROM metric WHERE id = $1 AND m_type = 'gauge'`, metricID).
					Scan(&storedValue)
				require.NoError(t, err)
				assert.True(t, storedValue.Valid)
				assert.Equal(t, *expectedValue, storedValue.Float64)
			}
		})
	}
}
