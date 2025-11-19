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

func (s *MemStorage) Update(ctx context.Context, m model.Metric) (*model.Metric, error) {
	// todo: next sprint
	// в текущем спринте не дается никаких требований на хранение метрик
	// поэтому сейчас метрики типа Gauge перезатирают значение,
	// а метрики типа Counter инкрементируют значение.
	// Вероятно далее необходимо будет сохранять значение с конкретной
	// временной меткой, но в рамках 1-го спринта это избыточно.
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

func (s *MemStorage) List(_ context.Context) (map[string]model.Metric, error) {
	// todo: next sprints
	// Возвращает копию мапы с метриками - не самый оптимальный вариант,
	// Но т.к. требования к структуре хранения метрик вероятно будет
	// обновлено в следующих спринтах - для упрощения пока сделано так.
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.metrics), nil
}

func (s *MemStorage) Get(_ context.Context, metricType model.MetricType, metricName string) (*model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metrics[metricName]
	if !ok || m.MType != metricType {
		return nil, model.ErrMetricNotFound
	}
	return &m, nil
}

func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}
