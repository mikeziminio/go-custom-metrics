package main

import (
	"context"
	stdlog "log"

	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/server/config"
	"go.uber.org/zap"
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
	ms, err := memstorage.New(syncWithUpdate, c.Restore, c.FileStoragePath, logger)
	if err != nil {
		logger.Fatal("failed to init memstorage", zap.Error(err))
	}

	s := server.New(
		c.Address,
		c.StoreInterval,
		ms,
		logger,
	)
	s.RegisterRoutes()
	s.Run(ctx)
}
