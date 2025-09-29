package main

import (
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/config"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
)

func main() {
	logger := log.New()
	ms := memstorage.New()
	err := server.StartServer(config.Port, ms, logger)
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
