package main

import (
	"context"
	stdlog "log"
	"time"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/dbstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/server/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := config.NewFromEnvsAndFlags()
	if err != nil {
		stdlog.Fatalf("failed to init config: %v", err)
	}
	logger, err := log.New(c.LogLevel)
	if err != nil {
		stdlog.Fatalf("failed to init logger: %v", err)
	}

	var syncWithUpdate bool
	if c.StoreInterval == 0 {
		syncWithUpdate = true
	}

	var storage server.Storage
	if c.DatabaseDSN == "" {
		ms, err := memstorage.New(syncWithUpdate, c.FileStoragePath, logger)
		if err != nil {
			logger.Fatal("failed to init memstorage", zap.Error(err))
		}
		storage = ms
	} else {
		ds, err := dbstorage.New(c.DatabaseDSN, logger)
		if err != nil {
			logger.Fatal("failed to init dbstorage", zap.Error(err))
		}
		storage = ds
	}

	if c.Restore {
		syncer, ok := storage.(server.Syncer)
		if !ok {
			logger.Warn("failed to restore, can't assert storage type as syncer")
		} else {
			syncer.Restore(ctx)
		}
	}

	s := server.New(
		c.Address,
		time.Duration(float64(time.Second)*c.StoreInterval),
		storage,
		logger,
	)
	s.RegisterRoutes()
	s.Run(ctx)
}
