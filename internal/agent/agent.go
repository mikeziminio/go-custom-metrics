package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

var (
	MetricAlloc         = "Alloc"
	MetricBuckHashSys   = "BuckHashSys"
	MetricFrees         = "Frees"
	MetricGCCPUFraction = "GCCPUFraction"
	MetricGCSys         = "GCSys"
	MetricHeapAlloc     = "HeapAlloc"
	MetricHeapIdle      = "HeapIdle"
	MetricHeapInuse     = "HeapInuse"
	MetricHeapObjects   = "HeapObjects"
	MetricHeapReleased  = "HeapReleased"
	MetricHeapSys       = "HeapSys"
	MetricLastGC        = "LastGC"
	MetricLookups       = "Lookups"
	MetricMCacheInuse   = "MCacheInuse"
	MetricMCacheSys     = "MCacheSys"
	MetricMSpanInuse    = "MSpanInuse"
	MetricMSpanSys      = "MSpanSys"
	MetricMallocs       = "Mallocs"
	MetricNextGC        = "NextGC"
	MetricNumForcedGC   = "NumForcedGC"
	MetricNumGC         = "NumGC"
	MetricOtherSys      = "OtherSys"
	MetricPauseTotalNs  = "PauseTotalNs"
	MetricStackInuse    = "StackInuse"
	MetricStackSys      = "StackSys"
	MetricSys           = "Sys"
	MetricTotalAlloc    = "TotalAlloc"
	MetricPollCount     = "PollCount"
	MetricRandomValue   = "RandomValue"
)

type Agent struct {
	pollInterval   float64
	reportInterval float64
	gauges         map[string]float64
	counters       map[string]int64
	mu             sync.RWMutex
	client         *http.Client
	baseURL        string
	sem            *semaphore.Weighted
	logger         *zap.Logger
}

func New(
	baseURL string,
	pollInterval float64,
	reportInterval float64,
	concurrentRequests int,
	logger *zap.Logger,
) *Agent {
	client := &http.Client{}
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		gauges:         make(map[string]float64),
		counters:       make(map[string]int64),
		client:         client,
		baseURL:        baseURL,
		sem:            semaphore.NewWeighted(int64(concurrentRequests)),
		logger:         logger,
	}
}

func (a *Agent) Collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	a.mu.Lock()
	defer a.mu.Unlock()
	a.gauges[MetricAlloc] = float64(ms.Alloc)
	a.gauges[MetricBuckHashSys] = float64(ms.BuckHashSys)
	a.gauges[MetricFrees] = float64(ms.Frees)
	a.gauges[MetricGCCPUFraction] = ms.GCCPUFraction
	a.gauges[MetricGCSys] = float64(ms.GCSys)
	a.gauges[MetricHeapAlloc] = float64(ms.HeapAlloc)
	a.gauges[MetricHeapIdle] = float64(ms.HeapIdle)
	a.gauges[MetricHeapInuse] = float64(ms.HeapInuse)
	a.gauges[MetricHeapObjects] = float64(ms.HeapObjects)
	a.gauges[MetricHeapReleased] = float64(ms.HeapReleased)
	a.gauges[MetricHeapSys] = float64(ms.HeapSys)
	a.gauges[MetricLastGC] = float64(ms.LastGC)
	a.gauges[MetricLookups] = float64(ms.Lookups)
	a.gauges[MetricMCacheInuse] = float64(ms.MCacheInuse)
	a.gauges[MetricMCacheSys] = float64(ms.MCacheSys)
	a.gauges[MetricMSpanInuse] = float64(ms.MSpanInuse)
	a.gauges[MetricMSpanSys] = float64(ms.MSpanSys)
	a.gauges[MetricMallocs] = float64(ms.Mallocs)
	a.gauges[MetricNextGC] = float64(ms.NextGC)
	a.gauges[MetricNumForcedGC] = float64(ms.NumForcedGC)
	a.gauges[MetricNumGC] = float64(ms.NumGC)
	a.gauges[MetricOtherSys] = float64(ms.OtherSys)
	a.gauges[MetricPauseTotalNs] = float64(ms.PauseTotalNs)
	a.gauges[MetricStackInuse] = float64(ms.StackInuse)
	a.gauges[MetricStackSys] = float64(ms.StackSys)
	a.gauges[MetricSys] = float64(ms.Sys)
	a.gauges[MetricTotalAlloc] = float64(ms.TotalAlloc)
	a.gauges[MetricRandomValue] = rand.Float64() //nolint:gosec // it's ok
	a.counters[MetricPollCount]++
}

func (a *Agent) Send(ctx context.Context, m *model.Metric) error {
	a.logger.Info("send metric start", zap.String("metric", fmt.Sprintf("%v", m)))
	var u string
	var err error
	switch m.MType {
	case model.Counter:
		u, err = url.JoinPath(
			a.baseURL, "/update",
			fmt.Sprintf("/%s/%s/%d", string(m.MType), m.ID, *m.Delta),
		)
	case model.Gauge:
		u, err = url.JoinPath(
			a.baseURL, "/update",
			fmt.Sprintf("/%s/%s/%.5f", string(m.MType), m.ID, *m.Value),
		)
	default:
		panic(fmt.Sprintf("unknown metric type: %s", m.MType))
	}
	if err != nil {
		return fmt.Errorf("failed to join url path for sending metric %s, %v", a.baseURL, m)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to init request: %w", err)
	}
	_ = a.sem.Acquire(ctx, 1)
	defer a.sem.Release(1)
	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close() //nolint:errcheck // it's ok
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for request %s: %d", u, res.StatusCode)
	}
	a.logger.Info("sent metric successfully", zap.String("url", u))

	return nil
}

func (a *Agent) SendAll(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(len(a.gauges))
	for name, val := range a.gauges {
		go func() {
			defer wg.Done()
			m := model.Metric{
				ID:    name,
				MType: model.Gauge,
				Value: &val,
			}
			err := a.Send(ctx, &m)
			if err != nil {
				a.logger.Fatal("failed to send metric", zap.Error(err))
			}
		}()
	}
	wg.Add(len(a.counters))
	for name, val := range a.counters {
		go func() {
			defer wg.Done()
			m := model.Metric{
				ID:    name,
				MType: model.Counter,
				Delta: &val,
			}
			err := a.Send(ctx, &m)
			if err != nil {
				a.logger.Fatal("failed to send metric", zap.Error(err))
			}
		}()
	}
	wg.Wait()
}

func (a *Agent) Run(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond * time.Duration(a.pollInterval*1000))
			a.Collect()
		}
	}()

	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond * time.Duration(a.reportInterval*1000))
			a.SendAll(ctx)
		}
	}()

	wg.Wait()
}
