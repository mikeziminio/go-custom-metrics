---
name: go-prometheus
description: Comprehensive Prometheus observability skill for Go projects, enabling robust metrics collection and monitoring integration
---

# Go Prometheus Observability Skill

## Overview
Comprehensive Prometheus observability skill for Go projects, enabling robust metrics collection and monitoring integration.

## Description
This skill provides comprehensive Prometheus observability capabilities for Go projects, including metric definition, collection, exposition, and integration with monitoring systems.

## Roles
- senior devops engineer
- senior go backend engineer

## Dependencies
- go 1.26+
- prometheus/client_golang
- prometheus/client_model
- prometheus/common

## Prometheus Integration Patterns

### 1. Basic Metric Definition
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Counter metrics
var (
    httpRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})

    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint"})
)

// Gauge metrics
var (
    activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of active connections",
    })
    
    memoryUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "memory_usage_bytes",
        Help: "Memory usage in bytes",
    }, []string{"process"})
)
```

### 2. HTTP Handler Instrumentation
```go
import (
    "net/http"
    "time"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Middleware for HTTP metrics
func instrumentHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Record request
        httpRequestCount.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
        
        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        defer func() {
            duration := time.Since(start).Seconds()
            httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
        }()
        
        next.ServeHTTP(wrapped, r)
    })
}

// Response writer wrapper to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// Register metrics endpoint
func registerMetricsHandler() {
    http.Handle("/metrics", promhttp.Handler())
}
```

### 3. Application-Level Metrics
```go
// Custom collector for application-specific metrics
type AppCollector struct {
    requestsTotal *prometheus.CounterVec
    errorsTotal   *prometheus.CounterVec
}

func NewAppCollector() *AppCollector {
    return &AppCollector{
        requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
            Name: "app_requests_total",
            Help: "Total application requests",
        }, []string{"service", "operation"}),
        errorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
            Name: "app_errors_total",
            Help: "Total application errors",
        }, []string{"service", "error_type"}),
    }
}

func (ac *AppCollector) IncRequest(service, operation string) {
    ac.requestsTotal.WithLabelValues(service, operation).Inc()
}

func (ac *AppCollector) IncError(service, errorType string) {
    ac.errorsTotal.WithLabelValues(service, errorType).Inc()
}
```

### 4. Database Metrics
```go
// Database connection pool metrics
var (
    dbConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "db_connections",
        Help: "Database connections",
    }, []string{"state"})
    
    dbQueries = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "db_queries_total",
        Help: "Total database queries",
    }, []string{"query_type", "status"})
)

// Wrap database connection
func wrapDBConnection(db *sql.DB) *sql.DB {
    // Set up connection pool metrics
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // Monitor pool stats
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            stats := db.Stats()
            dbConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
            dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))
        }
    }()
    
    return db
}
```

### 5. Service-Level Metrics
```go
// Service health metrics
var (
    serviceStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "service_status",
        Help: "Service status (1 = healthy, 0 = unhealthy)",
    }, []string{"service"})
    
    serviceLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "service_latency_seconds",
        Help: "Service latency in seconds",
        Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
    }, []string{"service", "endpoint"})
)

// Service health check
func (s *Service) HealthCheck() bool {
    // Perform health check
    healthy := s.performHealthCheck()
    
    if healthy {
        serviceStatus.WithLabelValues(s.name).Set(1)
    } else {
        serviceStatus.WithLabelValues(s.name).Set(0)
    }
    
    return healthy
}
```

## Taskfile Integration

Add the following to your root Taskfile.yml:

```yaml
# Prometheus metrics tasks
metrics-collect:
  desc: Collect and expose metrics
  cmds:
    - echo "Starting metrics collection..."
    - go run cmd/main.go
    - echo "Metrics exposed on :8080/metrics"

metrics-test:
  desc: Test metrics endpoints
  cmds:
    - echo "Testing metrics endpoints..."
    - curl -f localhost:8080/metrics || echo "Metrics endpoint not ready"
    - echo "Metrics test completed"

metrics-verify:
  desc: Verify metrics consistency
  cmds:
    - echo "Verifying metrics..."
    - go test -v ./internal/metrics
    - echo "Metrics verification completed"
```

## Prometheus Configuration

### prometheus.yml Configuration
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'go-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

## Best Practices

### 1. Metric Naming Conventions
```go
// Good naming
var (
    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
    }, []string{"method", "endpoint"})
    
    userServiceRequests = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "user_service_requests_total",
        Help: "Total user service requests",
    }, []string{"action", "status"})
)
```

### 2. Label Usage Guidelines
- Use labels for dimensions that vary frequently
- Avoid high-cardinality labels
- Keep label names consistent across metrics
- Limit number of labels per metric (typically 3-5)

### 3. Histogram Bucket Selection
```go
// For different use cases
var (
    // API response times
    apiResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "api_response_time_seconds",
        Help: "API response time in seconds",
        Buckets: prometheus.DefBuckets, // Default buckets
    }, []string{"endpoint"})
    
    // Latency-sensitive operations
    fastOperationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "fast_operation_latency_seconds",
        Help: "Fast operation latency in seconds",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    }, []string{"operation"})
)
```

## Sample Implementation

### Basic Service with Prometheus Metrics
```go
package main

import (
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})
    
    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint"})
)

func main() {
    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    
    // Start HTTP server
    http.ListenAndServe(":8080", nil)
}
```

## Monitoring Dashboards

### Grafana Dashboard Example
```json
{
  "title": "Go Service Metrics",
  "panels": [
    {
      "title": "Requests Per Second",
      "targets": [
        {
          "expr": "rate(http_requests_total[1m])",
          "legendFormat": "{{method}} {{endpoint}}"
        }
      ]
    },
    {
      "title": "Request Duration",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[1m])) by (le))",
          "legendFormat": "95th percentile"
        }
      ]
    }
  ]
}
```

## Error Handling and Metrics
```go
// Report errors through metrics
func processRequest(r *http.Request) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    }()
    
    // Process request
    err := doWork()
    if err != nil {
        // Log error and increment error counter
        httpRequestCount.WithLabelValues(r.Method, r.URL.Path, "500").Inc()
        return err
    }
    
    // Success
    httpRequestCount.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
    return nil
}
```

## Testing Metrics

### Prometheus Metrics Test Example
```go
func TestMetrics(t *testing.T) {
    // Reset metrics before test
    httpRequestCount.Reset()
    
    // Test metric collection
    req, _ := http.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Simulate request processing
        httpRequestCount.WithLabelValues("GET", "/test", "200").Inc()
    })
    
    handler.ServeHTTP(w, req)
    
    // Verify metrics
    assert.Equal(t, 1.0, httpRequestCount.WithLabelValues("GET", "/test", "200").Get())
}
```
