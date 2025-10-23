package log

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	defaultStatusCode int
	written           bool
	bytesWritten      int
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(rw.defaultStatusCode)
	}
	rw.bytesWritten += len(data)
	return rw.ResponseWriter.Write(data)
}

type LoggerMiddleware struct {
	logger *zap.Logger
}

func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

func (m *LoggerMiddleware) MiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, defaultStatusCode: http.StatusOK}

		m.logger.Info("Request started",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)

		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		m.logger.Info("Request completed",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status_code", wrapped.defaultStatusCode),
			zap.Int("bytes_written", wrapped.bytesWritten),
			zap.Duration("duration", duration),
		)
	})
}
