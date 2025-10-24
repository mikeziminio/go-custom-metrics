package main

import (
	"context"
	stdlog "log"

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
	ms := memstorage.New(syncWithUpdate, c.FileStoragePath, logger)

	s := server.New(
		c.Address,
		c.StoreInterval,
		c.Restore,
		ms,
		logger,
	)
	s.RegisterRoutes()
	s.Run(ctx)
}
