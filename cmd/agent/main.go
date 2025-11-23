package main

import (
	"context"
	"fmt"
	stdlog "log"
	"time"

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
		time.Duration(float64(time.Second)*c.PollInterval),
		time.Duration(float64(time.Second)*c.ReportInterval),
		c.UseCompress,
		time.Duration(float64(time.Second)*c.Timeout),
		logger,
	)

	a.Run(ctx)
}
