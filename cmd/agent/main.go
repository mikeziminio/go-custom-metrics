package main

import (
	"context"
	"fmt"
	stdlog "log"

	"github.com/mikeziminio/go-custom-metrics/internal/agent"
	"github.com/mikeziminio/go-custom-metrics/internal/agent/config"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
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

	a := agent.New(
		fmt.Sprintf("http://%s", c.Address),
		c.PollInterval,
		c.ReportInterval,
		c.UseCompress,
		logger,
	)

	a.Run(ctx)
}
