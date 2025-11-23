package syncer

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

const (
	StorageFileMode = 0600
)

type FileSyncer struct {
	fileStoragePath string
	logger          *zap.Logger
}

func New(fileStoragePath string, logger *zap.Logger) *FileSyncer {
	return &FileSyncer{
		fileStoragePath: fileStoragePath,
		logger: logger.With(
			zap.String("fileStoragePath", fileStoragePath),
		),
	}
}

func (s *FileSyncer) Restore() ([]model.Metric, error) {
	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to restore from file: %w", err)
	}

	var metricSlice []model.Metric

	err = json.Unmarshal(data, &metricSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return metricSlice, nil
}

func (s *FileSyncer) Sync(metricSlice []model.Metric) error {
	data, err := json.Marshal(metricSlice)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.WriteFile(s.fileStoragePath, data, StorageFileMode)
	if err != nil {
		return fmt.Errorf("failed to write %d bytes to %s",
			len(data), s.fileStoragePath)
	}

	return nil
}
