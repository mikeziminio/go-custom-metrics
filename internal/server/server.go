package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/compress"
	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

type Storage interface {
	Update(m model.Metric) (*model.Metric, error)
	List() map[string]model.Metric
	Get(metricType model.MetricType, metricName string) (*model.Metric, error)
	Sync() error
	Restore() error
}

// todo: next sprints
// Из задания 1-го спринта:
// Хендлеры должны взаимодействовать с экземпляром MemStorage при помощи соответствующих интерфейсных методов.
//
// Соответственно сейчас так и реализовано - без слоев service и repository, их использование планируется
// в следующих спринтах.

type APIServer struct {
	address       string
	storeInterval time.Duration
	restore       bool
	storage       Storage
	router        *chi.Mux
	httpServer    *http.Server
	logger        *zap.Logger
}

func New(
	address string,
	storeInterval float64,
	restore bool,
	storage Storage,
	logger *zap.Logger,
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
		storeInterval: time.Duration(float64(time.Second) * storeInterval),
		restore:       restore,
		storage:       storage,
		router:        r,
		httpServer:    httpServer,
		logger:        logger,
	}

	return a
}

func (a *APIServer) RegisterRoutes() {
	r := a.router

	lmw := log.NewLoggerMiddleware(a.logger)

	r.Use(middleware.StripSlashes)
	r.Use(lmw.MiddlewareHandler)
	r.Use(compress.DecompressMiddlewareHandler)
	r.Use(compress.CompressMiddlewareHandler)

	r.Get("/", a.List)
	r.Post("/value", a.Get)
	r.Get("/value/{metricType}/{metricName}", a.GetByParams)
	r.Post("/update", a.Update)
	r.Post("/update/{metricType}/{metricName}/{value}", a.UpdateByParams)
}

func (a *APIServer) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if a.restore {
		err := a.storage.Restore()
		if err != nil {
			a.logger.Fatal("failed to restore storage",
				zap.Error(err),
			)
		}
	}

	go func() {
		a.logger.Info("Server started", zap.String("address", a.httpServer.Addr))
		err := a.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	if a.storeInterval != 0 {
		go func() {
			t := time.NewTicker(a.storeInterval)

			a.logger.Info("File sync started",
				zap.Duration("storeInterval", a.storeInterval),
			)
			for {
				select {
				case <-t.C:
					err := a.storage.Sync()
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-ctx.Done():
	}

	err := a.httpServer.Shutdown(context.Background())
	if err != nil {
		a.logger.Fatal("failed to gracefully shutdown", zap.Error(err))
	}
	a.logger.Info("Server stopped")
}
