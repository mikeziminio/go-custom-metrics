package memstorage

import (
	"maps"
	"sync"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
)

type MemStorage struct {
	metrics map[string]model.Metric
	mu      sync.RWMutex
}

var _ server.Storage = (*MemStorage)(nil)

func New() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]model.Metric),
	}
}

func (s *MemStorage) Update(m model.Metric) error {
	// todo: next sprint
	// в текущем спринте не дается никаких требований на хранение метрик
	// поэтому сейчас метрики типа Gauge перезатирают значение,
	// а метрики типа Counter инкрементируют значение.
	// Вероятно далее необходимо будет сохранять значение с конкретной
	// временной меткой, но в рамках 1-го спринта это избыточно.
	s.mu.Lock()
	defer s.mu.Unlock()
	// s.metrics = append(s.metrics, m)
	current, ok := s.metrics[m.ID]
	if ok && m.MType == model.Counter {
		*m.Delta += *current.Delta
	}
	s.metrics[m.ID] = m
	return nil
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
	if !ok {
		return nil, model.ErrMetricNotFound
	}
	return &m, nil
}
