// Package server provides the HTTP API server for the metrics service.
//
// It implements the main APIServer struct with routes for metric operations.
// It supports both in-memory and database-backed storage through interfaces.
package server

import (
	"context"
	"net/http"

	// #nosec G108
	// Profiling endpoint is intentionally exposed on /debug/pprof
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/compress"
	"github.com/mikeziminio/go-custom-metrics/internal/hasher"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

// Storage interface defines the contract for metric storage implementations.
//
// Implementations must support update, retrieval, and listing operations
// for metrics with proper concurrency safety.
type Storage interface {
	// Update updates a single metric and returns the updated value.
	Update(ctx context.Context, m model.Metric) (*model.Metric, error)
	// Updates updates multiple metrics in bulk.
	Updates(ctx context.Context, metrics []model.Metric) error
	// List returns all metrics as a map.
	List(ctx context.Context) (map[string]model.Metric, error)
	// Get retrieves a specific metric by type and name.
	Get(ctx context.Context, metricType model.MetricType, metricName string) (*model.Metric, error)
	// Ping checks if the storage is reachable.
	Ping(ctx context.Context) error
}

// Syncer interface defines methods for file synchronization.
//
// Implementations should handle persistence of metrics to file.
type Syncer interface {
	// Sync saves current metrics state to file.
	Sync(ctx context.Context) error
	// Restore loads metrics state from file.
	Restore(ctx context.Context) error
}

// todo: next sprints
// Из задания 1-го спринта:
// Хендлеры должны взаимодействовать с экземпляром MemStorage при помощи соответствующих интерфейсных методов.
//
// Соответственно сейчас так и реализовано - без слоев service и repository, их использование планируется
// в следующих спринтах.

// APIServer represents the main HTTP server for the metrics API.
//
// It manages routes, request processing, and background synchronization tasks.
type APIServer struct {
	address       string
	storeInterval time.Duration
	hashKey       []byte
	storage       Storage
	router        *chi.Mux
	httpServer    *http.Server
	logger        *zap.Logger
	auditLogger   *AuditLogger
	pprofAddress  string
}

// New creates a new APIServer instance.
//
// Parameters:
//   - address: HTTP server address (host:port)
//   - storeInterval: Interval for file synchronization (0 to disable)
//   - hashKey: Key for request body hashing (nil to disable)
//   - storage: Storage implementation (memstorage or dbstorage)
//   - logger: Logger instance
//   - auditLogger: Audit logger instance (nil to disable auditing)
//   - pprofAddress: Address for pprof profiling server (empty to disable)
//
// Returns a new APIServer ready for route registration and startup.
func New(
	address string,
	storeInterval time.Duration,
	hashKey []byte,
	storage Storage,
	logger *zap.Logger,
	auditLogger *AuditLogger,
	pprofAddress string,
) *APIServer {
	r := chi.NewRouter()

	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	a := &APIServer{
		address:       address,
		storeInterval: storeInterval,
		hashKey:       hashKey,
		storage:       storage,
		router:        r,
		httpServer:    httpServer,
		logger:        logger,
		auditLogger:   auditLogger,
		pprofAddress:  pprofAddress,
	}

	return a
}

// RegisterRoutes registers all HTTP routes for the metrics API.
func (a *APIServer) RegisterRoutes() {
	r := a.router

	r.Use(middleware.StripSlashes)
	r.Use(log.MiddlewareHandler(a.logger))
	r.Use(compress.DecompressMiddlewareHandler)
	r.Use(hasher.MiddlewareHandler(a.hashKey, a.logger))
	r.Use(compress.CompressMiddlewareHandler)

	r.Get("/", a.List)
	r.Get("/ping", a.Ping)
	r.Post("/value", a.Get)
	r.Get("/value/{metricType}/{metricName}", a.GetByParams)
	r.Post("/update", a.Update)
	r.Post("/update/{metricType}/{metricName}/{value}", a.UpdateByParams)
	r.Post("/updates", a.Updates)
}

// Run starts the HTTP server and background tasks.
//
// It starts:
//   - HTTP server on configured address
//   - pprof server (if configured)
//   - File sync goroutine (if storeInterval > 0)
//
// The function blocks until SIGINT or SIGTERM is received, then gracefully
// shuts down the server.
func (a *APIServer) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start pprof server with proper timeouts
	go func() {
		pprofAddr := a.pprofAddress
		a.logger.Info("Starting pprof server", zap.String("address", pprofAddr))
		pprofServer := &http.Server{
			Addr:         pprofAddr,
			Handler:      nil, // DefaultServeMux от pprof
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		err := pprofServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.logger.Error("failed to start pprof server", zap.Error(err))
		}
	}()

	go func() {
		a.logger.Info("Server started", zap.String("address", a.httpServer.Addr))
		err := a.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	if a.storeInterval != 0 {
		syncer, ok := a.storage.(Syncer)
		if !ok {
			a.logger.Warn("failed to sync, can't assert storage type as syncer")
		} else {
			go func() {
				t := time.NewTicker(a.storeInterval)
				a.logger.Info("File sync started",
					zap.Duration("storeInterval", a.storeInterval),
				)
				for {
					select {
					case <-t.C:
						err := syncer.Sync(ctx)
						if err != nil {
							// судя по тому как сделаны тесты yandex - в случае ошибки синхронизации
							// сервер не должен убиваться
							a.logger.Warn("Failed to sync with file", zap.Error(err))
						}
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	ctx, cancel = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	err := a.httpServer.Shutdown(context.Background())
	if err != nil {
		a.logger.Fatal("failed to gracefully shutdown", zap.Error(err))
	}
	a.logger.Info("Server stopped")
}
