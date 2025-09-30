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
	ms := memstorage.New()

	s := server.New(c.Address, ms, logger)
	s.RegisterRoutes()
	s.Run(ctx)
}
