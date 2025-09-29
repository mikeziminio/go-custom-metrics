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
