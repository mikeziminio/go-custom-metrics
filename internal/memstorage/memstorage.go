package memstorage

import (
	"sync"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
)

type MemStorage struct {
	metrics []model.Metric
	mu      sync.RWMutex
}

var _ server.Storage = (*MemStorage)(nil)

func New() *MemStorage {
	return &MemStorage{}
}

func (s *MemStorage) Add(m model.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = append(s.metrics, m)
	return nil
}
