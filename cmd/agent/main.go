package main

import (
	"context"

	"github.com/mikeziminio/go-custom-metrics/internal/agent"
	"github.com/mikeziminio/go-custom-metrics/internal/config"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := log.New()
	a := agent.New(
		config.ServerBaseURL,
		config.PollInterval,
		config.ReportInterval,
		config.ConcurrentRequests,
		logger,
	)
	a.Run(ctx)
}
