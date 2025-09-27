package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

var updateRe = regexp.MustCompile(`^/update/(\w+)/(\w+)/(\w+)/?$`)

// Update implement handler for /update/<metricType>/<metricName>/<value>/
func (a *API) Update(res http.ResponseWriter, req *http.Request) {
	a.logger.Info("request start", zap.String("path", req.URL.Path))
	els := updateRe.FindStringSubmatch(req.URL.Path)

	if len(els) == 0 {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	metricType, err := model.NewMetricTypeFromString(els[1])
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	metricName := els[2]
	var delta *int64
	var value *float64

	if metricType == model.Counter {
		d, err := strconv.ParseInt(els[3], 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		delta = &d
	}

	if metricType == model.Gauge {
		v, err := strconv.ParseFloat(els[3], 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		value = &v
	}

	err = a.storage.Add(model.Metric{
		ID:    metricName,
		MType: metricType,
		Delta: delta,
		Value: value,
		Hash:  "",
	})
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	a.logger.Info("request end", zap.String("els", fmt.Sprintf("%#v", els)))
	res.WriteHeader(http.StatusOK)
}
