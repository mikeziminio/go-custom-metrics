package memstorage

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func (s *MemStorage) syncLogger() *zap.Logger {
	return s.logger.With(
		zap.String("fileStoragePath", s.fileStoragePath),
	)
}

func (s *MemStorage) restore() error {
	logger := s.syncLogger()
	logger.Info("Start restore from file")

	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		return fmt.Errorf("failed to restore from file: %w", err)
	}

	var metricList []model.Metric

	err = json.Unmarshal(data, &metricList)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	metricMap := make(map[string]model.Metric)
	for _, m := range metricList {
		metricMap[m.ID] = m
	}
	s.metrics = metricMap
	logger.Info("Finish restore from file")
	return nil
}

func (s *MemStorage) Sync() error {
	logger := s.syncLogger()
	logger.Info("Start sync with file")

	s.mu.RLock()
	values := maps.Values(s.metrics)
	s.mu.RUnlock()

	res := slices.Collect(values)
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.WriteFile(s.fileStoragePath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write %d bytes to %s",
			len(data), s.fileStoragePath)
	}

	logger.Info("Finish sync with file")
	return nil
}
