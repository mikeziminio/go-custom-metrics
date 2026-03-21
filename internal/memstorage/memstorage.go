package memstorage

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/syncer"
)

type MemStorage struct {
	metrics        map[string]model.Metric
	mu             sync.RWMutex
	syncWithUpdate bool
	syncer         *syncer.FileSyncer
	logger         *zap.Logger
}

var _ server.Storage = (*MemStorage)(nil)

// New creates a new MemStorage instance.
//
// Parameters:
//   - syncWithUpdate: Whether to sync to file after each update
//   - fileStoragePath: Path to the file for synchronization
//   - logger: Logger instance
//
// Returns a new MemStorage and an error if file syncer initialization fails.
func New(syncWithUpdate bool, fileStoragePath string, logger *zap.Logger) (*MemStorage, error) {
	snc := syncer.New(fileStoragePath, logger)
	s := MemStorage{
		syncWithUpdate: syncWithUpdate,
		syncer:         snc,
		metrics:        make(map[string]model.Metric),
		logger:         logger,
	}
	return &s, nil
}

// Update updates a single metric in memory.
//
// For counters, the value is incremented. For gauges, the value is replaced.
// If syncWithUpdate is true, it also syncs to file.
func (s *MemStorage) Update(ctx context.Context, m model.Metric) (*model.Metric, error) {
	// todo: next sprint
	// в текущем спринте не дается никаких требований на хранение метрик
	// поэтому сейчас метрики типа Gauge перезатирают значение,
	// а метрики типа Counter инкрементируют значение.
	s.mu.Lock()
	current, ok := s.metrics[m.ID]
	if ok && m.MType == model.Counter {
		*m.Delta += *current.Delta
	}
	s.metrics[m.ID] = m
	s.mu.Unlock()

	if s.syncWithUpdate {
		err := s.Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to sync storage: %w", err)
		}
	}
	return &m, nil
}

// Updates updates multiple metrics in memory.
//
// For counters, the values are incremented. For gauges, the values are replaced.
// If syncWithUpdate is true, it also syncs to file.
func (s *MemStorage) Updates(ctx context.Context, metrics []model.Metric) error {
	s.mu.Lock()
	for _, m := range metrics {
		current, ok := s.metrics[m.ID]
		if ok && m.MType == model.Counter {
			*m.Delta += *current.Delta
		}
		s.metrics[m.ID] = m
	}
	s.mu.Unlock()

	if s.syncWithUpdate {
		err := s.Sync(ctx)
		if err != nil {
			return fmt.Errorf("failed to sync storage: %w", err)
		}
	}
	return nil
}

// List returns a copy of all metrics from memory.
func (s *MemStorage) List(_ context.Context) (map[string]model.Metric, error) {
	// todo: next sprints
	// Возвращает копию мапы с метриками - не самый оптимальный вариант,
	// Но т.к. требования к структуре хранения метрик вероятно будет
	// обновлено в следующих спринтах - для упрощения пока сделано так.
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.metrics), nil
}

// Get retrieves a specific metric from memory.
//
// Returns ErrMetricNotFound if the metric doesn't exist or has a different type.
func (s *MemStorage) Get(_ context.Context, metricType model.MetricType, metricName string) (*model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metrics[metricName]
	if !ok || m.MType != metricType {
		return nil, model.ErrMetricNotFound
	}
	return &m, nil
}

// Ping checks if the storage is alive (always returns nil).
func (*MemStorage) Ping(_ context.Context) error {
	return nil
}
