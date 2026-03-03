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

// AuditLogger handles logging audit events to file and/or HTTP endpoint
type AuditLogger struct {
	logger *zap.Logger
	file   *os.File
	client *http.Client
	url    string
	mu     sync.RWMutex
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
	}

	return al, nil
}

// Log logs an audit event to configured destinations
func (al *AuditLogger) Log(ctx context.Context, event AuditEvent) error {
	var wg sync.WaitGroup
	var errs []error

	// Log to file if configured
	if al.file != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := al.logToFile(event); err != nil {
				errs = append(errs, fmt.Errorf("file logging failed: %w", err))
			}
		}()
	}

	// Log to HTTP endpoint if configured
	if al.url != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := al.logToHTTP(ctx, event); err != nil {
				errs = append(errs, fmt.Errorf("HTTP logging failed: %w", err))
			}
		}()
	}

	wg.Wait()

	// Combine all errors
	if len(errs) > 0 {
		return fmt.Errorf("audit logging errors: %v", errs)
	}

	return nil
}

// logToFile writes audit event to file
func (al *AuditLogger) logToFile(event AuditEvent) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	_, err = al.file.WriteString(string(data) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

// logToHTTP sends audit event to HTTP endpoint
func (al *AuditLogger) logToHTTP(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", al.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := al.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	return nil
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
