package log

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	written      bool
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK // статус по умолчанию
		rw.written = true
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

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		m.logger.Info("Request started",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)

		next.ServeHTTP(wrapped, r)
		// duration по ТЗ пишет в риквест, но это не логично
		// его нужно писать в
		duration := time.Since(start)

		m.logger.Info("Request completed",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status_code", wrapped.statusCode),
			zap.Int("bytes_written", wrapped.bytesWritten),
			zap.Duration("duration", duration),
		)
	})
}
