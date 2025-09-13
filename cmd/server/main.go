package main

import (
	"log"

	"github.com/mikeziminio/go-custom-metrics/internal/config"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
)

func main() {
	ms := memstorage.New()
	err := server.StartServer(config.Port, ms)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
