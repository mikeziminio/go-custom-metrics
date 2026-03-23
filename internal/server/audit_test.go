package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestAuditLogger_Register(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() (*AuditLogger, *MockObserver)
		expectedCount int
	}{
		{
			name: "register single observer",
			setup: func() (*AuditLogger, *MockObserver) {
				logger := zap.L()
				config := AuditConfig{}
				auditLogger, _ := NewAuditLogger(logger, config)
				mockObserver := NewMockObserver(t)
				auditLogger.Register(mockObserver)
				return auditLogger, mockObserver
			},
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger, _ := tc.setup()
			auditLogger.mu.RLock()
			assert.Len(t, auditLogger.observers, tc.expectedCount)
			auditLogger.mu.RUnlock()
		})
	}
}

func TestAuditLogger_Deregister(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() (*AuditLogger, *MockObserver)
		expectedCount int
	}{
		{
			name: "deregister existing observer",
			setup: func() (*AuditLogger, *MockObserver) {
				logger := zap.L()
				config := AuditConfig{}
				auditLogger, _ := NewAuditLogger(logger, config)
				mockObserver := NewMockObserver(t)
				auditLogger.Register(mockObserver)
				auditLogger.Deregister(mockObserver)
				return auditLogger, mockObserver
			},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger, _ := tc.setup()
			auditLogger.mu.RLock()
			assert.Len(t, auditLogger.observers, tc.expectedCount)
			auditLogger.mu.RUnlock()
		})
	}
}

func TestAuditLogger_Notify(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		setup           func() (*AuditLogger, []*MockObserver)
		event           AuditEvent
		expectNoError   bool
		expectObservers []*MockObserver
	}{
		{
			name: "notify multiple observers",
			setup: func() (*AuditLogger, []*MockObserver) {
				logger := zap.L()
				config := AuditConfig{}
				auditLogger, _ := NewAuditLogger(logger, config)
				mockObserver1 := NewMockObserver(t)
				mockObserver2 := NewMockObserver(t)

				// Set up expectations for the mock observers
				mockObserver1.EXPECT().Update(mock.AnythingOfType("server.AuditEvent")).Return(nil)
				mockObserver2.EXPECT().Update(mock.AnythingOfType("server.AuditEvent")).Return(nil)

				auditLogger.Register(mockObserver1)
				auditLogger.Register(mockObserver2)
				return auditLogger, []*MockObserver{mockObserver1, mockObserver2}
			},
			event: AuditEvent{
				Timestamp: 1234567890,
				Metrics:   []string{"metric1", "metric2"},
				IPAddress: "127.0.0.1",
			},
			expectNoError:   true,
			expectObservers: []*MockObserver{{}, {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger, observers := tc.setup()
			err := auditLogger.Notify(tc.event)

			if tc.expectNoError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			// Verify that all observers were called
			for _, observer := range observers {
				observer.AssertExpectations(t)
			}
		})
	}
}

func TestAuditLogger_Log(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() (*AuditLogger, *MockObserver, AuditEvent)
		expectNoError bool
	}{
		{
			name: "log event successfully",
			setup: func() (*AuditLogger, *MockObserver, AuditEvent) {
				logger := zap.L()
				config := AuditConfig{}
				auditLogger, _ := NewAuditLogger(logger, config)
				mockObserver := NewMockObserver(t)

				// Set up expectation for the mock observer
				mockObserver.EXPECT().Update(mock.AnythingOfType("server.AuditEvent")).Return(nil)

				auditLogger.Register(mockObserver)

				event := AuditEvent{
					Timestamp: 1234567890,
					Metrics:   []string{"metric1", "metric2"},
					IPAddress: "127.0.0.1",
				}
				return auditLogger, mockObserver, event
			},
			expectNoError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger, _, event := tc.setup()
			ctx := context.Background()
			err := auditLogger.Log(ctx, event)

			if tc.expectNoError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuditLogger_NewAuditLogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		config                AuditConfig
		expectedObserverCount int
		expectNoError         bool
	}{
		{
			name: "file logging only",
			config: AuditConfig{
				AuditFile: "/tmp/test-audit.log",
			},
			expectedObserverCount: 1,
			expectNoError:         true,
		},
		{
			name: "http logging only",
			config: AuditConfig{
				AuditURL: "http://localhost:8080/audit",
			},
			expectedObserverCount: 1,
			expectNoError:         true,
		},
		{
			name: "both logging types",
			config: AuditConfig{
				AuditFile: "/tmp/test-audit.log",
				AuditURL:  "http://localhost:8080/audit",
			},
			expectedObserverCount: 2,
			expectNoError:         true,
		},
		{
			name:                  "no logging configured",
			config:                AuditConfig{},
			expectedObserverCount: 0,
			expectNoError:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file for testing (if needed)
			if tc.config.AuditFile != "" {
				tmpFile, err := os.CreateTemp("", "audit-test-*.log")
				assert.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				tc.config.AuditFile = tmpFile.Name()
			}

			logger := zap.L()
			auditLogger, err := NewAuditLogger(logger, tc.config)

			if tc.expectNoError {
				assert.NoError(t, err)
				assert.NotNil(t, auditLogger)
			} else {
				assert.Error(t, err)
			}

			if tc.expectedObserverCount > 0 {
				auditLogger.mu.RLock()
				assert.Len(t, auditLogger.observers, tc.expectedObserverCount)
				auditLogger.mu.RUnlock()
			}

			// Clean up if needed
			if auditLogger != nil {
				auditLogger.Close()
			}
		})
	}
}

func TestFileObserver_Update(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setup         func() (*FileObserver, string)
		event         AuditEvent
		expectNoError bool
	}{
		{
			name: "update file observer successfully",
			setup: func() (*FileObserver, string) {
				tmpFile, err := os.CreateTemp("", "audit-file-test-*.log")
				assert.NoError(t, err)
				tmpFileName := tmpFile.Name()
				tmpFile.Close()

				fileObserver := &FileObserver{file: nil} // Will be replaced
				return fileObserver, tmpFileName
			},
			event: AuditEvent{
				Timestamp: 1234567890,
				Metrics:   []string{"metric1", "metric2"},
				IPAddress: "127.0.0.1",
			},
			expectNoError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileObserver, tmpFileName := tc.setup()

			// Use filepath.Abs to prevent path traversal
			absPath, err := filepath.Abs(tmpFileName)
			if err != nil {
				t.Fatal(err)
			}

			// Open file for the observer
			file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			assert.NoError(t, err)
			defer file.Close()
			defer os.Remove(tmpFileName) // Move defer here to clean up after test

			fileObserver.file = file

			err = fileObserver.Update(tc.event)

			if tc.expectNoError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			if tc.expectNoError {
				// Verify file content
				content, err := os.ReadFile(tmpFileName)
				assert.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

func TestHTTPObserver_Update(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setup       func() *HTTPObserver
		event       AuditEvent
		expectError bool
	}{
		{
			name: "update HTTP observer with invalid URL",
			setup: func() *HTTPObserver {
				// We'll create an HTTP observer that connects to a non-existent URL
				// This will cause an error during the update, which is the expected behavior
				// for this test case
				return &HTTPObserver{
					client: &http.Client{Timeout: 1 * time.Second},
					url:    "http://invalid-url-that-does-not-exist-12345.com/audit",
				}
			},
			event: AuditEvent{
				Timestamp: 1234567890,
				Metrics:   []string{"metric1", "metric2"},
				IPAddress: "127.0.0.1",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpObserver := tc.setup()
			err := httpObserver.Update(tc.event)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
