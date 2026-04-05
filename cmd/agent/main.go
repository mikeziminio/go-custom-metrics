package main

import (
	"context"
	"fmt"
	stdlog "log"
	"time"

	"github.com/mikeziminio/go-custom-metrics/internal/agent"
	"github.com/mikeziminio/go-custom-metrics/internal/agent/config"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit,
	)

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
	logger.Info("config", zap.Any("config", c))

	a := agent.New(
		fmt.Sprintf("http://%s", c.Address),
		time.Duration(float64(time.Second)*c.PollInterval),
		time.Duration(float64(time.Second)*c.ReportInterval),
		c.UseCompress,
		[]byte(c.HashKey),
		c.RateLimit,
		time.Duration(float64(time.Second)*c.Timeout),
		logger,
		c.CryptoKey,
	)

	a.Run(ctx)
}
