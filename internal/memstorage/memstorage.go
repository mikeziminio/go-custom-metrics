package memstorage

import (
	"fmt"
	"maps"
	"sync"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"go.uber.org/zap"
)

type MemStorage struct {
	metrics         map[string]model.Metric
	mu              sync.RWMutex
	syncWithUpdate  bool
	fileStoragePath string
	logger          *zap.Logger
}

var _ server.Storage = (*MemStorage)(nil)

func New(syncWithUpdate bool, fileStoragePath string, logger *zap.Logger) *MemStorage {
	return &MemStorage{
		syncWithUpdate:  syncWithUpdate,
		fileStoragePath: fileStoragePath,
		metrics:         make(map[string]model.Metric),
		logger:          logger,
	}
}

func (s *MemStorage) Update(m model.Metric) (*model.Metric, error) {
	// todo: next sprint
	// в текущем спринте не дается никаких требований на хранение метрик
	// поэтому сейчас метрики типа Gauge перезатирают значение,
	// а метрики типа Counter инкрементируют значение.
	// Вероятно далее необходимо будет сохранять значение с конкретной
	// временной меткой, но в рамках 1-го спринта это избыточно.
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.metrics[m.ID]
	if ok && m.MType == model.Counter {
		*m.Delta += *current.Delta
	}
	s.metrics[m.ID] = m
	if s.syncWithUpdate {
		err := s.Sync()
		if err != nil {
			return nil, fmt.Errorf("failed to sync storage")
		}
	}
	return &m, nil
}

func (s *MemStorage) List() map[string]model.Metric {
	// todo: next sprints
	// Возвращает копию мапы с метриками - не самый оптимальный вариант,
	// Но т.к. требования к структуре хранения метрик вероятно будет
	// обновлено в следующих спринтах - для упрощения пока сделано так.
	s.mu.RLock()
	defer s.mu.RUnlock()
	return maps.Clone(s.metrics)
}

func (s *MemStorage) Get(metricType model.MetricType, metricName string) (*model.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metrics[metricName]
	if !ok || m.MType != metricType {
		return nil, model.ErrMetricNotFound
	}
	return &m, nil
}
