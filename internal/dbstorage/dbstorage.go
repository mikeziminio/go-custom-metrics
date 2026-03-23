// Package dbstorage provides database-backed storage for metrics.
//
// It implements the storage layer using PostgreSQL with support for
// counters and gauges. The package includes automatic migration support
// and retry mechanisms for database operations.
package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/retrier"
)

// DBStorage provides database-backed storage for metrics using PostgreSQL.
//
// It supports two metric types:
//   - Counter: Incremental values stored as delta
//   - Gauge: Point-in-time values stored as current value
//
// DBStorage implements automatic retry logic for database operations
// and uses SQL migrations for schema management.
type DBStorage struct {
	db      *sql.DB
	retrier Retrier
	logger  *zap.Logger
}

// Retrier interface defines the retry mechanism contract.
//
// Implementations wrap a function with automatic retry logic,
// classifying errors to determine which should trigger a retry.
type Retrier interface {
	Retry(f func() error) error
}

var defaultRetryTimeouts = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// New creates a new DBStorage instance and applies database migrations.
//
// It establishes a connection to PostgreSQL, sets up retry logic,
// and runs all pending migrations.
//
// Parameters:
//   - connString: PostgreSQL connection string
//   - logger: Logger instance for logging
//
// Returns a new DBStorage or an error if connection or migration fails.
func New(connString string, logger *zap.Logger) (*DBStorage, error) {
	var db *sql.DB
	r := retrier.NewRetrier(defaultRetryTimeouts, newRetryClassifier())
	err := r.Retry(func() (e error) {
		db, e = sql.Open("pgx", connString)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	s := DBStorage{
		db:      db,
		retrier: r,
		logger:  logger,
	}
	err = s.migrateUp(connString)
	if err != nil {
		logger.Fatal("failed to migrate up", zap.Error(err))
	}
	return &s, nil
}

func (s *DBStorage) migrateUp(connString string) error {
	// Для миграций - отдельное подключение
	var db *sql.DB
	err := s.retrier.Retry(func() (e error) {
		db, e = sql.Open("pgx", connString)
		return
	})
	if err != nil {
		return fmt.Errorf("failed to create connection pool for migration: %w", err)
	}
	defer db.Close()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current filename")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsPath := "file://" + filepath.Join(projectRoot, "migrations")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath, "postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate up: %w", err)
	}

	return nil
}

// Update stores a single metric in the database using a transaction.
//
// For counters, it increments the delta. For gauges, it updates the value.
// The operation is wrapped in a transaction with automatic retry logic.
//
// Parameters:
//   - ctx: Context for the database operation
//   - m: Metric to store (must have valid ID, MType, and either Delta or Value)
//
// Returns the stored metric with updated values or an error.
func (s *DBStorage) Update(ctx context.Context, m model.Metric) (*model.Metric, error) {
	var tx *sql.Tx
	err := s.retrier.Retry(func() (e error) {
		tx, e = s.db.BeginTx(ctx, nil)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var um *model.Metric
	var e error
	switch m.MType {
	case model.Counter:
		um, e = s.updateCounter(ctx, tx, m)
	case model.Gauge:
		um, e = s.updateGauge(ctx, tx, m)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", m.MType)
	}
	if e != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", e)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return um, nil
}

// createUniqueMetrics подготавливает слайс метрик с уникальными ID для Updates
func createUniqueMetrics(metrics []model.Metric) []model.Metric {
	// Для корректной обработки дубликатов:
	// - для counter: суммируем delta
	// - для gauge: заменяем value
	seenCounters := make(map[string]int64)
	seenGauges := make(map[string]float64)

	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			if m.Delta != nil {
				if existingDelta, exists := seenCounters[m.ID]; exists {
					seenCounters[m.ID] = existingDelta + *m.Delta
				} else {
					seenCounters[m.ID] = *m.Delta
				}
			}
		case model.Gauge:
			if m.Value != nil {
				seenGauges[m.ID] = *m.Value
			}
		}
	}

	uniqueMetrics := make([]model.Metric, 0, len(seenCounters)+len(seenGauges))

	for id, delta := range seenCounters {
		uniqueMetrics = append(uniqueMetrics, model.Metric{
			ID:    id,
			MType: model.Counter,
			Delta: &delta,
			Value: nil,
		})
	}

	for id, value := range seenGauges {
		uniqueMetrics = append(uniqueMetrics, model.Metric{
			ID:    id,
			MType: model.Gauge,
			Delta: nil,
			Value: &value,
		})
	}

	return uniqueMetrics
}

// Updates stores multiple metrics in a single batch operation.
//
// It handles duplicates by:
//   - Summing deltas for counters
//   - Replacing values for gauges
//
// The operation uses PostgreSQL's ON CONFLICT upsert pattern.
//
// Parameters:
//   - ctx: Context for the database operation
//   - metrics: Slice of metrics to store
//
// Returns an error if the batch operation fails.
func (s *DBStorage) Updates(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// мы должны подготовить уникальные по ID метрики,
	// иначе upsert вернет ошибку
	// SQLSTATE 21000: ON CONFLICT DO UPDATE command cannot affect row a second time
	uniqueMetrics := createUniqueMetrics(metrics)

	query := `INSERT INTO metric (id, m_type, delta, value) VALUES`
	var args []any
	for i, m := range uniqueMetrics {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4) // #nolint:mnd
		args = append(args, m.ID, m.MType, m.Delta, m.Value)
	}
	query += `
		ON CONFLICT (id) DO UPDATE SET
			delta = CASE WHEN EXCLUDED.m_type = 'counter' THEN metric.delta + EXCLUDED.delta ELSE metric.delta END,
			value = CASE WHEN EXCLUDED.m_type = 'gauge' THEN EXCLUDED.value ELSE metric.value END
	`

	err := s.retrier.Retry(func() (e error) {
		_, e = s.db.ExecContext(ctx, query, args...)
		return
	})
	if err != nil {
		return fmt.Errorf("failed to upsert metrics: %w", err)
	}

	return nil
}

func (s *DBStorage) updateCounter(ctx context.Context, tx *sql.Tx, m model.Metric) (*model.Metric, error) {
	err := s.retrier.Retry(func() (e error) {
		_, e = tx.ExecContext(ctx,
			`INSERT INTO metric (id, m_type, delta) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET
			m_type = EXCLUDED.m_type,
			delta = metric.delta + EXCLUDED.delta`,
			m.ID, m.MType, m.Delta,
		)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert counter: %w", err)
	}

	var newDelta sql.NullInt64
	err = s.retrier.Retry(func() (e error) {
		e = tx.QueryRowContext(ctx,
			`SELECT delta FROM metric WHERE id = $1 AND m_type = $2`,
			m.ID, m.MType,
		).Scan(&newDelta)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated counter: %w", err)
	}

	if !newDelta.Valid {
		return nil, fmt.Errorf("failed to update delta, it's null for %s", m.ID)
	}
	updatedDelta := newDelta.Int64

	return &model.Metric{
		ID:    m.ID,
		MType: m.MType,
		Delta: &updatedDelta,
	}, nil
}

func (s *DBStorage) updateGauge(ctx context.Context, tx *sql.Tx, m model.Metric) (*model.Metric, error) {
	err := s.retrier.Retry(func() (e error) {
		_, e = tx.ExecContext(ctx,
			`INSERT INTO metric (id, m_type, value) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET
			m_type = EXCLUDED.m_type,
			value = EXCLUDED.value`,
			m.ID, m.MType, m.Value,
		)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert gauge: %w", err)
	}

	return &model.Metric{
		ID:    m.ID,
		MType: m.MType,
		Value: m.Value,
	}, nil
}

func (s *DBStorage) Close() error {
	if s.db != nil {
		s.db.Close()
	}

	return nil
}

// Ping tests the database connection health.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns an error if the database is unreachable.
func (s *DBStorage) Ping(ctx context.Context) error {
	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// Get retrieves a single metric from the database by ID and type.
//
// Parameters:
//   - ctx: Context for the database operation
//   - metricType: Type of metric (counter or gauge)
//   - metricName: ID of the metric to retrieve
//
// Returns the metric if found, or model.ErrMetricNotFound if not present.
func (s *DBStorage) Get(ctx context.Context, metricType model.MetricType, metricName string) (*model.Metric, error) {
	var m model.Metric
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := s.retrier.Retry(func() (e error) {
		e = s.db.QueryRowContext(
			ctx,
			`SELECT id, m_type, delta, value FROM metric WHERE id = $1 AND m_type = $2`,
			metricName,
			metricType,
		).
			Scan(&m.ID, &m.MType, &delta, &value)
		return
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrMetricNotFound
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	if m.MType == model.Counter {
		if delta.Valid {
			m.Delta = &delta.Int64
		}
	} else {
		if value.Valid {
			m.Value = &value.Float64
		}
	}

	return &m, nil
}

// List retrieves all metrics from the database.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns a map of metrics keyed by ID, or an error if the query fails.
func (s *DBStorage) List(ctx context.Context) (map[string]*model.Metric, error) {
	var rows *sql.Rows
	err := s.retrier.Retry(func() (e error) {
		rows, e = s.db.QueryContext(ctx, "SELECT id, m_type, delta, value FROM metric")
		// rows.Err() проверяется ниже. Здесь добавлено, чтобы проходила проверка statictest
		// т.к. директивы аналогичной //nolint для него нет
		if false {
			_ = rows.Err()
		}
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics from database: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string]*model.Metric)
	for rows.Next() {
		var m model.Metric
		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}

		if m.MType == model.Counter {
			if delta.Valid {
				m.Delta = &delta.Int64
			}
		} else {
			if value.Valid {
				m.Value = &value.Float64
			}
		}

		metrics[m.ID] = &m
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return metrics, nil
}
