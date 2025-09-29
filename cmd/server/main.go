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

	server := server.New(c.Address, ms, logger)
	server.Run(ctx)
}
