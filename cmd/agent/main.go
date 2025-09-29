package main

import (
	"context"
	"fmt"

	"github.com/mikeziminio/go-custom-metrics/internal/agent"
	"github.com/mikeziminio/go-custom-metrics/internal/agent/config"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
)

func main() {
	c, _ := config.NewFromFlags()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := log.New()
	a := agent.New(
		fmt.Sprintf("http://%s", c.Address),
		c.PollInterval,
		c.ReportInterval,
		c.ConcurrentRequests,
		logger,
	)
	a.Run(ctx)
}
