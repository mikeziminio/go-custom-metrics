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
	"github.com/mikeziminio/go-custom-metrics/internal/syncer"
)

type DBStorage struct {
	db             *sql.DB
	connString     string
	syncWithUpdate bool
	syncer         *syncer.Syncer
	retrier        Retrier
	logger         *zap.Logger
}

var defaultRetryTimeouts = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func New(connString string, syncWithUpdate bool, fileStoragePath string, logger *zap.Logger) (*DBStorage, error) {
	var db *sql.DB
	retrier := NewPgRetrier(defaultRetryTimeouts)
	err := retrier.Retry(func() (e error) {
		db, e = sql.Open("pgx", connString)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	snc := syncer.New(fileStoragePath, logger)
	s := DBStorage{
		db:             db,
		connString:     connString,
		syncWithUpdate: syncWithUpdate,
		syncer:         snc,
		retrier:        retrier,
		logger:         logger,
	}
	return &s, nil
}

func (s *DBStorage) MigrateUp() error {
	// Для миграций - отдельное подключение
	var db *sql.DB
	err := s.retrier.Retry(func() (e error) {
		db, e = sql.Open("pgx", s.connString)
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

func (s *DBStorage) Updates(ctx context.Context, metrics []model.Metric) error {
	var tx *sql.Tx
	err := s.retrier.Retry(func() (e error) {
		tx, e = s.db.BeginTx(ctx, nil)
		return
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, m := range metrics {
		var e error
		switch m.MType {
		case model.Counter:
			_, e = s.updateCounter(ctx, tx, m)
		case model.Gauge:
			_, e = s.updateGauge(ctx, tx, m)
		default:
			return fmt.Errorf("unsupported metric type: %s", m.MType)
		}
		if e != nil {
			return fmt.Errorf("failed to update transaction: %w", e)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// updateCounter handles updating counter metrics with proper accumulation
func (s *DBStorage) updateCounter(ctx context.Context, tx *sql.Tx, m model.Metric) (*model.Metric, error) {
	// Проверим есть ли уже такой счетчик
	var existingDelta sql.NullInt64
	err := s.retrier.Retry(func() (e error) {
		e = tx.QueryRowContext(ctx,
			`SELECT delta FROM metric WHERE id = $1 AND m_type = $2`,
			m.ID, m.MType,
		).Scan(&existingDelta)
		return
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to query row: %w", err)
		}
		// Счетчика нет, добавим новый
		err := s.retrier.Retry(func() (e error) {
			_, e = tx.ExecContext(ctx,
				`INSERT INTO metric (id, m_type, delta) VALUES ($1, $2, $3)`,
				m.ID, m.MType, m.Delta,
			)
			return
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert new counter: %w", err)
		}
		return &model.Metric{
			ID:    m.ID,
			MType: m.MType,
			Delta: m.Delta,
		}, nil
	}

	// Счетчик есть, пересчитаем значение
	var newDelta int64
	if existingDelta.Valid {
		newDelta = existingDelta.Int64 + *m.Delta
	} else {
		newDelta = *m.Delta
	}
	err = s.retrier.Retry(func() (e error) {
		_, e = tx.ExecContext(ctx,
			`UPDATE metric SET delta = $1 WHERE id = $2 AND m_type = $3`,
			newDelta, m.ID, m.MType,
		)
		return
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update counter: %w", err)
	}

	return &model.Metric{
		ID:    m.ID,
		MType: m.MType,
		Delta: &newDelta,
	}, nil
}

// updateGauge handles updating gauge metrics by replacing the value
func (s *DBStorage) updateGauge(ctx context.Context, tx *sql.Tx, m model.Metric) (*model.Metric, error) {
	// Проверим, есть ли уже значение
	var existingFloat sql.NullFloat64
	err := s.retrier.Retry(func() (e error) {
		e = tx.QueryRowContext(ctx,
			`SELECT value FROM metric WHERE id = $1 AND m_type = $2`,
			m.ID, m.MType,
		).Scan(&existingFloat)
		return
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to query row: %w", err)
		}
		// Значения нет, добавим новое
		err := s.retrier.Retry(func() (e error) {
			_, e = tx.ExecContext(ctx,
				`INSERT INTO metric (id, m_type, value) VALUES ($1, $2, $3)`,
				m.ID, m.MType, m.Value,
			)
			return
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert new gauge: %w", err)
		}
	} else {
		// Значение есть, обновим
		err := s.retrier.Retry(func() (e error) {
			_, err = tx.ExecContext(ctx,
				`UPDATE metric SET value = $1 WHERE id = $2 AND m_type = $3`,
				m.Value, m.ID, m.MType,
			)
			return
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update gauge: %w", err)
		}
	}

	return &model.Metric{
		ID:    m.ID,
		MType: m.MType,
		Value: m.Value,
	}, nil
}

// Close закрывает соединение с базой данных
func (s *DBStorage) Close() error {
	// Close database connection
	if s.db != nil {
		s.db.Close()
	}

	return nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	// Используем контекст с операцией пинга
	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

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
		if err == sql.ErrNoRows {
			return nil, model.ErrMetricNotFound
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	// Set the appropriate value field based on metric type
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

func (s *DBStorage) List(ctx context.Context) (map[string]model.Metric, error) {
	// Retrieve all metrics from database
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

	metrics := make(map[string]model.Metric)
	for rows.Next() {
		var m model.Metric
		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}

		// Set the appropriate value field based on metric type
		if m.MType == model.Counter {
			if delta.Valid {
				m.Delta = &delta.Int64
			}
		} else {
			if value.Valid {
				m.Value = &value.Float64
			}
		}

		metrics[m.ID] = m
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return metrics, nil
}
