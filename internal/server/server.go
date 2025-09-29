package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

type Storage interface {
	Update(m model.Metric) error
	List() map[string]model.Metric
	Get(metricType model.MetricType, metricName string) (*model.Metric, error)
}

type API struct {
	storage Storage
	logger  *zap.Logger
}

func StartServer(address string, storage Storage, logger *zap.Logger) error {
	api := &API{
		storage: storage,
		logger:  logger,
	}

	r := chi.NewRouter()
	r.Get("/", api.List)
	r.Get("/value/{metricType}/{metricName}", api.Get)
	r.Post("/update/{metricType}/{metricName}/{value}", api.Update)

	err := http.ListenAndServe(address, r)
	return err
}
