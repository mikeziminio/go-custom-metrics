package memstorage

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func (s *MemStorage) Restore(_ context.Context) error {
	metricSlice, err := s.syncer.Restore()
	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	metricMap := make(map[string]model.Metric)
	for _, m := range metricSlice {
		metricMap[m.ID] = m
	}
	s.metrics = metricMap
	return nil
}

func (s *MemStorage) Sync(_ context.Context) error {
	s.mu.RLock()
	values := maps.Values(s.metrics)
	s.mu.RUnlock()
	metricSlice := slices.Collect(values)

	err := s.syncer.Sync(metricSlice)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}
	return nil
}
