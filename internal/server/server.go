package server

import (
	"context"
	"net/http"
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
	Update(ctx context.Context, m model.Metric) (*model.Metric, error)
	Updates(ctx context.Context, metrics []model.Metric) error
	List(ctx context.Context) (map[string]model.Metric, error)
	Get(ctx context.Context, metricType model.MetricType, metricName string) (*model.Metric, error)
	Ping(ctx context.Context) error
}

type Syncer interface {
	Sync(ctx context.Context) error
	Restore(ctx context.Context) error
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
	storage       Storage
	router        *chi.Mux
	httpServer    *http.Server
	logger        *zap.Logger
}

func New(
	address string,
	storeInterval time.Duration,
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
		storeInterval: storeInterval,
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
	r.Get("/ping", a.Ping)
	r.Post("/value", a.Get)
	r.Get("/value/{metricType}/{metricName}", a.GetByParams)
	r.Post("/update", a.Update)
	r.Post("/update/{metricType}/{metricName}/{value}", a.UpdateByParams)
	r.Post("/updates", a.Updates)
}

func (a *APIServer) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
