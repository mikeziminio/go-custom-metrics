package dbstorage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func (s *DBStorage) Restore(ctx context.Context) error {
	metrics, err := s.syncer.Restore()
	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	// Insert all metrics into database
	for _, m := range metrics {
		switch m.MType {
		case model.Counter:
			err = s.retrier.Retry(func() (e error) {
				_, e = s.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, delta)
					VALUES ($1, $2, $3)
					ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta`,
					m.ID, m.MType, m.Delta)
				return
			})
		case model.Gauge:
			err = s.retrier.Retry(func() (e error) {
				_, e = s.db.ExecContext(ctx, `
					INSERT INTO metric (id, m_type, value)
					VALUES ($1, $2, $3)
					ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value`,
					m.ID, m.MType, m.Value)
				return
			})
		default:
			return fmt.Errorf("unsupported metric type: %s", m.MType)
		}
		if err != nil {
			return fmt.Errorf("failed to insert metric %s: %w", m.ID, err)
		}
	}

	return nil
}

func (s *DBStorage) Sync(ctx context.Context) error {
	// Retrieve all metrics from database
	var rows *sql.Rows
	err := s.retrier.Retry(func() (e error) {
		rows, e = s.db.QueryContext(ctx, "SELECT id, m_type, delta, value FROM metric") //nolint // check for rows.Err() below
		// rows.Err() проверяется ниже. Здесь добавлено, чтобы проходила проверка statictest
		// т.к. директивы аналогичной //nolint для него нет
		if false {
			_ = rows.Err()
		}
		return
	})
	if err != nil {
		return fmt.Errorf("failed to query metrics from database: %w", err)
	}
	defer rows.Close()

	var metrics []model.Metric
	for rows.Next() {
		var m model.Metric
		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return fmt.Errorf("failed to scan metric row: %w", err)
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

		metrics = append(metrics, m)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	err = s.syncer.Sync(metrics)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}
	return nil
}
