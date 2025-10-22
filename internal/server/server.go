package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/log"
	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

type Storage interface {
	Update(m model.Metric) error
	List() map[string]model.Metric
	Get(metricType model.MetricType, metricName string) (*model.Metric, error)
}

// todo: next sprints
// Из задания 1-го спринта:
// Хендлеры должны взаимодействовать с экземпляром MemStorage при помощи соответствующих интерфейсных методов.
//
// Соответственно сейчас так и реализовано - без слоев service и repository, их использование планируется
// в следующих спринтах.

type APIServer struct {
	storage    Storage
	router     *chi.Mux
	httpServer *http.Server
	logger     *zap.Logger
}

func New(address string, storage Storage, logger *zap.Logger) *APIServer {
	r := chi.NewRouter()

	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	a := &APIServer{
		storage:    storage,
		router:     r,
		httpServer: httpServer,
		logger:     logger,
	}

	return a
}

func (a *APIServer) RegisterRoutes() {
	r := a.router

	lmw := log.NewLoggerMiddleware(a.logger)

	r.Use(lmw.MiddlewareHandler)

	r.Get("/", a.List)
	r.Get("/value/{metricType}/{metricName}", a.Get)
	r.Post("/update/{metricType}/{metricName}/{value}", a.Update)
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
