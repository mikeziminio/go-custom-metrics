package server

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

type Storage interface {
	Add(m model.Metric) error
}

type API struct {
	storage Storage
	logger  *zap.Logger
}

func StartServer(port int, storage Storage, logger *zap.Logger) error {
	api := &API{
		storage: storage,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", api.Update)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux) //nolint:gosec // no timeout
	return err
}
