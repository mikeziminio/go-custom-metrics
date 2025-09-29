package main

import (
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/server/config"
)

func main() {
	c, _ := config.NewFromFlags()
	logger := log.New()
	ms := memstorage.New()
	err := server.StartServer(c.Address, ms, logger)
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
