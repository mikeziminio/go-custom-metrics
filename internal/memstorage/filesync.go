package memstorage

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"go.uber.org/zap"
)

func (s *MemStorage) Restore() error {
	s.logger.Info("Start restore from file",
		zap.String("fileStoragePath", s.fileStoragePath),
	)
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		return nil
		// return fmt.Errorf("failed to read from %s", s.fileStoragePath)
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
	return nil
}

func (s *MemStorage) Sync() error {
	s.logger.Info("Start sync with file",
		zap.String("fileStoragePath", s.fileStoragePath),
	)
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := slices.Collect(maps.Values(s.metrics))
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.WriteFile(s.fileStoragePath, data, 0755)
	if err != nil {
		return fmt.Errorf("failed to write %d bytes to %s",
			len(data), s.fileStoragePath)
	}

	return nil
}
