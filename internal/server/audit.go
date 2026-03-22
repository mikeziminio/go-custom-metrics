package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AuditEvent represents a single audit event.
//
// It contains information about the metric update: timestamp, affected metrics,
// and the IP address of the client.
type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// AuditConfig holds configuration for audit logging.
//
// It specifies the destinations for audit events: file and/or HTTP endpoint.
type AuditConfig struct {
	AuditFile string
	AuditURL  string
}

// Observer interface for audit logging observers.
//
// Implementations receive audit events and process them (e.g., write to file, send HTTP).
type Observer interface {
	Update(event AuditEvent) error
}

// Subject interface for audit logging subject.
//
// It manages observers and notifies them about audit events.
type Subject interface {
	Register(observer Observer)
	Deregister(observer Observer)
	Notify(event AuditEvent) error
}

// AuditLogger implements both Subject and Observer interfaces.
//
// It manages multiple audit destinations (file and/or HTTP) and broadcasts
// audit events to all registered observers.
type AuditLogger struct {
	logger    *zap.Logger
	observers []Observer
	mu        sync.RWMutex
	client    *http.Client
	url       string
	file      *os.File
	filePath  string
}

// NewAuditLogger creates a new AuditLogger instance.
//
// Parameters:
//   - logger: Logger instance for logging audit operations
//   - config: Audit configuration with file and/or URL destinations
//
// Returns a new *AuditLogger and an error if file initialization fails.
func NewAuditLogger(logger *zap.Logger, config AuditConfig) (*AuditLogger, error) {
	al := &AuditLogger{
		logger: logger,
		client: &http.Client{Timeout: 5 * time.Second}, // #nolint:mnd
		url:    config.AuditURL,
	}

	if config.AuditFile != "" {
		file, err := os.OpenFile(config.AuditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit file %s: %w", config.AuditFile, err)
		}
		al.file = file
		al.filePath = config.AuditFile
	}

	// Register observers based on configuration
	if config.AuditFile != "" {
		fileObserver := &FileObserver{file: al.file}
		al.Register(fileObserver)
	}

	if config.AuditURL != "" {
		httpObserver := &HTTPObserver{client: al.client, url: al.url}
		al.Register(httpObserver)
	}

	return al, nil
}

// Register adds an observer to the audit logger.
//
// Parameters:
//   - observer: Observer instance to register
func (al *AuditLogger) Register(observer Observer) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.observers = append(al.observers, observer)
}

// Deregister removes an observer from the audit logger.
//
// Parameters:
//   - observer: Observer instance to remove
func (al *AuditLogger) Deregister(observer Observer) {
	al.mu.Lock()
	defer al.mu.Unlock()
	for i, obs := range al.observers {
		if obs == observer {
			al.observers = append(al.observers[:i], al.observers[i+1:]...)
			break
		}
	}
}

// Notify notifies all registered observers about an audit event.
//
// Parameters:
//   - event: AuditEvent to broadcast to all observers
//
// Returns an error if any observer fails to process the event.
func (al *AuditLogger) Notify(event AuditEvent) error {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var wg sync.WaitGroup
	var errs []error

	for _, observer := range al.observers {
		wg.Add(1)
		go func(obs Observer) {
			defer wg.Done()
			if err := obs.Update(event); err != nil {
				errs = append(errs, err)
			}
		}(observer)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("audit logging errors: %v", errs)
	}

	return nil
}

// Log logs an audit event to configured destinations.
//
// Parameters:
//   - ctx: Context for the operation (currently unused)
//   - event: AuditEvent to log
func (al *AuditLogger) Log(_ context.Context, event AuditEvent) error { // #nolint:revive
	return al.Notify(event)
}

// Close closes the audit logger and associated file.
//
// Returns an error if any close operation fails.
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	var errs []error

	if al.file != nil {
		if err := al.file.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close audit file: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("audit logger closing errors: %v", errs)
	}

	return nil
}

// FileObserver implements Observer interface for file logging.
type FileObserver struct {
	file *os.File
	mu   sync.RWMutex
}

// Update implements Observer interface for file logging.
//
// Parameters:
//   - event: AuditEvent to write to the file
//
// Returns an error if the event cannot be marshaled or written to the file.
//
// It appends the JSON-encoded audit event to the file.
func (fo *FileObserver) Update(event AuditEvent) error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	_, err = fo.file.WriteString(string(data) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

// HTTPObserver implements Observer interface for HTTP logging.
type HTTPObserver struct {
	client *http.Client
	url    string
}

// Update implements Observer interface for HTTP logging.
//
// Parameters:
//   - event: AuditEvent to send to the HTTP endpoint
//
// Returns an error if the request cannot be created or sent.
//
// It sends the JSON-encoded audit event to the HTTP endpoint.
func (ho *HTTPObserver) Update(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", ho.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ho.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	return nil
}

// extractIPAddress extracts the IP address from the request.
//
// It checks X-Forwarded-For, X-Real-IP headers, and falls back to RemoteAddr.
func extractIPAddress(r *http.Request) string {
	// First try to get IP from X-Forwarded-For header
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For might have multiple IPs, get the first one
		ips := []string{}
		for _, v := range []string{",", ";"} {
			ips = append(ips, strings.Split(ip, v)...)
		}
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
		// Validate IP format
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Then try X-Real-IP header
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Finally fall back to RemoteAddr
	if ip == "" {
		ip = r.RemoteAddr
		// Extract IP from "host:port" format
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		// Validate IP format
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// If we can't parse any IP, return empty string
	return ""
}
