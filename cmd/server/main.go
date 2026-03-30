package main

import (
	"context"
	"fmt"
	stdlog "log"

	// #nosec G108
	// Profiling endpoint is intentionally exposed on /debug/pprof
	_ "net/http/pprof"
	"time"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/dbstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/memstorage"
	"github.com/mikeziminio/go-custom-metrics/internal/server"
	"github.com/mikeziminio/go-custom-metrics/internal/server/config"
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

	var syncWithUpdate bool
	if c.StoreInterval == 0 {
		syncWithUpdate = true
	}

	var storage server.Storage
	if c.DatabaseDSN == "" {
		ms, err := memstorage.New(syncWithUpdate, c.FileStoragePath, logger)
		if err != nil {
			logger.Fatal("failed to init memstorage", zap.Error(err))
		}
		storage = ms
	} else {
		ds, err := dbstorage.New(c.DatabaseDSN, logger)
		if err != nil {
			logger.Fatal("failed to init dbstorage", zap.Error(err))
		}
		storage = ds
	}

	if c.Restore {
		syncer, ok := storage.(server.Syncer)
		if !ok {
			logger.Warn("failed to restore, can't assert storage type as syncer")
		} else {
			syncer.Restore(ctx)
		}
	}

	var auditLogger *server.AuditLogger
	if c.AuditFile != "" || c.AuditURL != "" {
		auditConfig := server.AuditConfig{
			AuditFile: c.AuditFile,
			AuditURL:  c.AuditURL,
		}
		auditLogger, err = server.NewAuditLogger(logger, auditConfig)
		if err != nil {
			logger.Fatal("failed to init audit logger", zap.Error(err))
		}
		defer func() {
			if auditLogger != nil {
				if err := auditLogger.Close(); err != nil {
					logger.Error("failed to close audit logger", zap.Error(err))
				}
			}
		}()
	}

	s := server.New(
		c.Address,
		time.Duration(float64(time.Second)*c.StoreInterval),
		[]byte(c.HashKey),
		storage,
		logger,
		auditLogger,
		c.PprofAddress,
	)
	s.RegisterRoutes()
	s.Run(ctx)
}
