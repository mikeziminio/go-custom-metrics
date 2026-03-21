// Package main is the entry point for the custom metrics agent application.
//
// The agent collects system metrics (CPU, memory, Go runtime statistics)
// and sends them to a configured metrics server for storage and retrieval.
//
// It supports concurrent metric collection and transmission with retry logic
// and compression for efficient network usage.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mikeziminio/go-custom-metrics/internal/agent"
	"github.com/mikeziminio/go-custom-metrics/internal/agent/config"
	zaplog "github.com/mikeziminio/go-custom-metrics/internal/log"
)

// main is the entry point for the custom metrics agent application.
//
// It initializes configuration, logger, and agent components then starts
// the metric collection and transmission loop.
//
// The agent runs until it receives SIGINT or SIGTERM signals.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := config.NewFromEnvsAndFlags()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}
	logger, err := zaplog.New(c.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	a := agent.New(
		fmt.Sprintf("http://%s", c.Address),
		time.Duration(float64(time.Second)*c.PollInterval),
		time.Duration(float64(time.Second)*c.ReportInterval),
		c.UseCompress,
		[]byte(c.HashKey),
		c.RateLimit,
		time.Duration(float64(time.Second)*c.Timeout),
		logger,
	)

	a.Run(ctx)
}
