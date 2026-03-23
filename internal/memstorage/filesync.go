package memstorage

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

// Restore loads metrics from the file into memory.
//
// Parameters:
//   - ctx: Context for the operation (ignored)
//
// Returns an error if the file cannot be read or parsed.
func (s *MemStorage) Restore(_ context.Context) error {
	metricSlice, err := s.syncer.Restore()
	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	metricMap := make(map[string]*model.Metric)
	for _, m := range metricSlice {
		metricMap[m.ID] = m
	}
	s.metrics = metricMap
	return nil
}

// Sync saves metrics from memory to the file.
//
// Parameters:
//   - ctx: Context for the operation (ignored)
//
// Returns an error if the file cannot be written.
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
