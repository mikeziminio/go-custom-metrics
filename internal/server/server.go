package server

import (
	"fmt"
	"net/http"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

type Storage interface {
	Add(m model.Metric) error
}

type API struct {
	storage Storage
}

func StartServer(port int, storage Storage) error {
	api := &API{
		storage: storage,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", api.Update)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux) //nolint:gosec // no timeout
	return err
}
