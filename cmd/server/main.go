package main

import (
	"context"

	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/server/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := config.NewFromFlags()
	logger := log.New()

	var syncWithUpdate bool
	if c.StoreInterval == 0 {
		syncWithUpdate = true
	}
	ms := memstorage.New(syncWithUpdate, c.FileStoragePath)

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
