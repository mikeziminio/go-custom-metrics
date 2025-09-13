package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

// POST /update/counter/someMetric/527
func (a *API) Update(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/update/")
	els := strings.Split(path, "/")

	if len(els) != 3 {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	metricType, err := model.NewMetricTypeFromString(els[0])
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricName := els[1]
	if metricName == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	val, err := strconv.ParseFloat(els[2], 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	err = a.storage.Add(model.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: nil,
		Value: &val,
		Hash:  "",
	})
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}
