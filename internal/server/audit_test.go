package server

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuditLogger_Register(t *testing.T) {
	logger := zap.L()
	config := AuditConfig{}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Create a mock observer
	mockObserver := &mockObserver{}

	// Register the observer
	auditLogger.Register(mockObserver)

	// Verify observer is registered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 1)
	auditLogger.mu.RUnlock()
}

func TestAuditLogger_Deregister(t *testing.T) {
	logger := zap.L()
	config := AuditConfig{}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Create a mock observer
	mockObserver := &mockObserver{}

	// Register the observer
	auditLogger.Register(mockObserver)

	// Verify observer is registered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 1)
	auditLogger.mu.RUnlock()

	// Deregister the observer
	auditLogger.Deregister(mockObserver)

	// Verify observer is deregistered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 0)
	auditLogger.mu.RUnlock()
}

func TestAuditLogger_Notify(t *testing.T) {
	logger := zap.L()
	config := AuditConfig{}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Create mock observers
	mockObserver1 := &mockObserver{}
	mockObserver2 := &mockObserver{}

	// Register observers
	auditLogger.Register(mockObserver1)
	auditLogger.Register(mockObserver2)

	// Create audit event
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	// Notify observers
	err = auditLogger.Notify(event)
	assert.NoError(t, err)

	// Verify both observers received the event
	assert.True(t, mockObserver1.receivedEvent)
	assert.True(t, mockObserver2.receivedEvent)
}

func TestAuditLogger_Log(t *testing.T) {
	logger := zap.L()
	config := AuditConfig{}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Create mock observer
	mockObserver := &mockObserver{}
	auditLogger.Register(mockObserver)

	// Create audit event
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	// Log the event
	ctx := context.Background()
	err = auditLogger.Log(ctx, event)
	assert.NoError(t, err)

	// Verify observer received the event
	assert.True(t, mockObserver.receivedEvent)
}

func TestAuditLogger_NewAuditLogger_FileLogging(t *testing.T) {
	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := zap.L()
	config := AuditConfig{
		AuditFile: tmpFile.Name(),
	}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Verify file observer was registered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 1)
	auditLogger.mu.RUnlock()

	// Clean up
	auditLogger.Close()
}

func TestAuditLogger_NewAuditLogger_HTTPLogging(t *testing.T) {
	logger := zap.L()
	config := AuditConfig{
		AuditURL: "http://localhost:8080/audit",
	}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Verify HTTP observer was registered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 1)
	auditLogger.mu.RUnlock()
}

func TestAuditLogger_NewAuditLogger_BothLogging(t *testing.T) {
	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := zap.L()
	config := AuditConfig{
		AuditFile: tmpFile.Name(),
		AuditURL:  "http://localhost:8080/audit",
	}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Verify both observers were registered
	auditLogger.mu.RLock()
	assert.Len(t, auditLogger.observers, 2)
	auditLogger.mu.RUnlock()

	// Clean up
	auditLogger.Close()
}

func TestFileObserver_Update(t *testing.T) {
	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "audit-file-test-*.log")
	assert.NoError(t, err)
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	logger := zap.L()
	config := AuditConfig{
		AuditFile: tmpFileName,
	}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Get the file observer
	auditLogger.mu.RLock()
	fileObserver := auditLogger.observers[0].(*FileObserver)
	auditLogger.mu.RUnlock()

	// Create audit event
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	// Update with event
	err = fileObserver.Update(event)
	assert.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(tmpFileName)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)

	// Clean up
	auditLogger.Close()
}

func TestHTTPObserver_Update(t *testing.T) {
	// Create a mock HTTP server to test the HTTP observer
	// We'll test error cases specifically
	logger := zap.L()
	config := AuditConfig{
		AuditURL: "http://invalid-url-that-does-not-exist-12345.com/audit",
	}
	auditLogger, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, auditLogger)

	// Get the HTTP observer
	auditLogger.mu.RLock()
	httpObserver := auditLogger.observers[0].(*HTTPObserver)
	auditLogger.mu.RUnlock()

	// Create audit event
	event := AuditEvent{
		Timestamp: 1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	// Update with event - should fail due to invalid URL
	err = httpObserver.Update(event)
	assert.Error(t, err)
}

// Mock observer for testing
type mockObserver struct {
	receivedEvent bool
	mu            sync.RWMutex
}

func (m *mockObserver) Update(event AuditEvent) error {
	m.mu.Lock()
	m.receivedEvent = true
	m.mu.Unlock()
	return nil
}

// Mock observer that returns an error for testing error handling
type errorObserver struct {
	updateError error
}

func (e *errorObserver) Update(event AuditEvent) error {
	return e.updateError
}
