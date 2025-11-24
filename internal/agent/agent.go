package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	mathrand "math/rand/v2"
	"net/http"
	"net/url"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	"github.com/mikeziminio/go-custom-metrics/internal/compress"
	"github.com/mikeziminio/go-custom-metrics/internal/hasher"
	"github.com/mikeziminio/go-custom-metrics/internal/model"
	"github.com/mikeziminio/go-custom-metrics/internal/retrier"
)

var (
	MetricAlloc          = "Alloc"
	MetricBuckHashSys    = "BuckHashSys"
	MetricFrees          = "Frees"
	MetricGCCPUFraction  = "GCCPUFraction"
	MetricGCSys          = "GCSys"
	MetricHeapAlloc      = "HeapAlloc"
	MetricHeapIdle       = "HeapIdle"
	MetricHeapInuse      = "HeapInuse"
	MetricHeapObjects    = "HeapObjects"
	MetricHeapReleased   = "HeapReleased"
	MetricHeapSys        = "HeapSys"
	MetricLastGC         = "LastGC"
	MetricLookups        = "Lookups"
	MetricMCacheInuse    = "MCacheInuse"
	MetricMCacheSys      = "MCacheSys"
	MetricMSpanInuse     = "MSpanInuse"
	MetricMSpanSys       = "MSpanSys"
	MetricMallocs        = "Mallocs"
	MetricNextGC         = "NextGC"
	MetricNumForcedGC    = "NumForcedGC"
	MetricNumGC          = "NumGC"
	MetricOtherSys       = "OtherSys"
	MetricPauseTotalNs   = "PauseTotalNs"
	MetricStackInuse     = "StackInuse"
	MetricStackSys       = "StackSys"
	MetricSys            = "Sys"
	MetricTotalAlloc     = "TotalAlloc"
	MetricPollCount      = "PollCount"
	MetricRandomValue    = "RandomValue"
	MetricTotalMemory    = "TotalMemory"
	MetricFreeMemory     = "FreeMemory"
	MetricCPUUtilization = "CPUutilization"
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	gauges         map[string]float64
	counters       map[string]int64
	mu             sync.RWMutex
	client         *http.Client
	baseURL        string
	useCompress    bool
	hashKey        []byte
	sem            *semaphore.Weighted
	retrier        Retrier
	logger         *zap.Logger
}

type Retrier interface {
	Retry(f func() error) error
}

var defaultRetryTimeouts = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// Ограничивает количество ретраев отправки метрик
// в зависимости от интервала отправки
func retryTimeouts(rts []time.Duration, reportInterval time.Duration) []time.Duration {
	var currentInterval time.Duration
	for i, timeout := range rts {
		currentInterval += timeout
		if currentInterval >= reportInterval {
			return rts[:i]
		}
	}
	return rts
}

func New(
	baseURL string,
	pollInterval time.Duration,
	reportInterval time.Duration,
	useCompress bool,
	hashKey []byte,
	rateLimit int,
	timeout time.Duration,
	logger *zap.Logger,
) *Agent {
	r := retrier.NewRetrier(
		retryTimeouts(defaultRetryTimeouts, reportInterval),
		retrier.NewDefaultRetryClassifier(),
	)
	client := &http.Client{
		Timeout: timeout,
	}
	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		gauges:         make(map[string]float64),
		counters:       make(map[string]int64),
		client:         client,
		baseURL:        baseURL,
		useCompress:    useCompress,
		hashKey:        hashKey,
		sem:            semaphore.NewWeighted(int64(rateLimit)),
		retrier:        r,
		logger:         logger,
	}
}

func randFloat64() float64 {
	b := make([]byte, 8) //nolint:mnd // 8 bytes for uint64
	_, err := rand.Read(b)
	if err != nil {
		return mathrand.Float64() //nolint:gosec // fallback
	}

	// Convert the bytes to a uint64
	val := binary.LittleEndian.Uint64(b)

	// Normalize the uint64 to a float64 in the range [0.0, 1.0)
	// by dividing by the maximum possible uint64 value plus 1.
	// This ensures a uniform distribution.
	return float64(val) / (float64(math.MaxUint64) + 1)
}

func (a *Agent) Collect() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		a.collectBasic()
	}()
	go func() {
		defer wg.Done()
		a.collectExtra()
	}()
	wg.Wait()
}

func (a *Agent) collectBasic() {
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
	a.gauges[MetricRandomValue] = randFloat64()
	a.counters[MetricPollCount]++
}

func (a *Agent) collectExtra() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		a.logger.Fatal("failed to fetch mem metrics")
	}
	utils, err := cpu.Percent(time.Second, true)
	if err != nil {
		a.logger.Fatal("failed to fetch cpu metrics")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.gauges[MetricTotalMemory] = float64(vm.Total)
	a.gauges[MetricFreeMemory] = float64(vm.Free)
	for i, u := range utils {
		a.gauges[MetricCPUUtilization+strconv.Itoa(i)] = u
	}
}

func (a *Agent) SendByBatch(ctx context.Context, metrics []model.Metric, useCompress bool) error {
	a.logger.Info("send metric start", zap.String("metric", fmt.Sprintf("%v", len(metrics))))

	u, err := url.JoinPath(a.baseURL, "/updates")
	if err != nil {
		return fmt.Errorf("failed to join url path for sending %d metrics, %v", len(metrics), a.baseURL)
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal %d metrics: %w", len(metrics), err)
	}

	var bodyReader io.Reader
	bodyReader = bytes.NewReader(body)
	if useCompress {
		bodyReader = compress.CompressWithGZIP(bodyReader)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to init request: %w", err)
	}
	if len(a.hashKey) > 0 {
		h := hasher.HexHash(body, a.hashKey)
		req.Header.Set(hasher.HashHeader, h)
	}
	req.Header.Set("Accept", "application/json")
	if useCompress {
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
	}

	err = a.retrier.Retry(func() (e error) {
		a.sem.Acquire(ctx, 1)
		res, e := a.client.Do(req)
		a.sem.Release(1)
		if e != nil {
			return e
		}
		defer res.Body.Close() //nolint:errcheck // it's ok
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code for request: %d", res.StatusCode)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	a.logger.Info("sent metric successfully",
		zap.Int("count", len(metrics)),
	)

	return nil
}

func (a *Agent) SendAll(ctx context.Context, useCompress bool) {
	metrics := make([]model.Metric, 0, len(a.gauges)+len(a.counters))
	for name, val := range a.gauges {
		metrics = append(metrics, model.Metric{
			ID:    name,
			MType: model.Gauge,
			Value: &val,
		})
	}
	for name, delta := range a.counters {
		metrics = append(metrics, model.Metric{
			ID:    name,
			MType: model.Counter,
			Delta: &delta,
		})
	}

	err := a.SendByBatch(ctx, metrics, useCompress)
	if err != nil {
		a.logger.Error("failed to send metrics by batch", zap.Error(err))
	}
}

func (a *Agent) Run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(a.pollInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.Collect()
			}
		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(a.reportInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.SendAll(ctx, a.useCompress)
			}
		}
	}()

	a.logger.Info("Agent started", zap.String("baseURL", a.baseURL))
	wg.Wait()
}
