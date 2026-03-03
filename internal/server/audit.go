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

// AuditEvent represents a single audit event
type AuditEvent struct {
	Timestamp int64    `json:"timestamp"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// AuditConfig holds configuration for audit logging
type AuditConfig struct {
	AuditFile string
	AuditURL  string
}

// Observer interface for audit logging observers
type Observer interface {
	Update(event AuditEvent) error
}

// Subject interface for audit logging subject
type Subject interface {
	Register(observer Observer)
	Deregister(observer Observer)
	Notify(event AuditEvent) error
}

// AuditLogger implements both Subject and Observer interfaces
type AuditLogger struct {
	logger    *zap.Logger
	observers []Observer
	mu        sync.RWMutex
	client    *http.Client
	url       string
	file      *os.File
	filePath  string
}

// NewAuditLogger creates a new AuditLogger instance
func NewAuditLogger(logger *zap.Logger, config AuditConfig) (*AuditLogger, error) {
	al := &AuditLogger{
		logger: logger,
		client: &http.Client{Timeout: 5 * time.Second},
		url:    config.AuditURL,
	}

	if config.AuditFile != "" {
		file, err := os.OpenFile(config.AuditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

// Register adds an observer to the audit logger
func (al *AuditLogger) Register(observer Observer) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.observers = append(al.observers, observer)
}

// Deregister removes an observer from the audit logger
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

// Notify notifies all registered observers about an audit event
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

// Log logs an audit event to configured destinations
func (al *AuditLogger) Log(ctx context.Context, event AuditEvent) error {
	return al.Notify(event)
}

// Close closes the audit logger and associated file
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

// FileObserver implements Observer interface for file logging
type FileObserver struct {
	file *os.File
	mu   sync.RWMutex
}

// Update implements Observer interface for file logging
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

// HTTPObserver implements Observer interface for HTTP logging
type HTTPObserver struct {
	client *http.Client
	url    string
}

// Update implements Observer interface for HTTP logging
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

// extractIPAddress extracts the IP address from the request
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
