package log

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
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

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		m.logger.Info("Request started",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("request_headers", fmt.Sprintf("%#v", r.Header)),
		)

		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		m.logger.Info("Request completed",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("response_headers", fmt.Sprintf("%#v", wrapped.Header())),
			zap.Int("status_code", wrapped.statusCode),
			zap.Int("bytes_written", wrapped.bytesWritten),
			zap.Duration("duration", duration),
		)
	})
}
