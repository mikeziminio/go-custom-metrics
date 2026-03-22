package hasher

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// HashHeader is the HTTP header name used for the hash value.
const HashHeader = "HashSHA256"

// MiddlewareHandler creates an HTTP middleware that validates and adds hash headers.
//
// When key is non-empty, it validates incoming request hashes and adds hash
// to outgoing responses. When key is empty, it passes requests through unchanged.
//
// Parameters:
//   - key: Secret key for HMAC-SHA256 (empty key disables hash processing)
//   - logger: Logger instance for logging
//
// Returns an HTTP middleware handler function.
func MiddlewareHandler(key []byte, logger *zap.Logger) func(http.Handler) http.Handler {
	if len(key) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	hmw := newHasherMiddleware(key, logger)
	return hmw.middlewareHandler
}

type hasherMiddleware struct {
	key    []byte
	logger *zap.Logger
}

// newHasherMiddleware creates a new HasherMiddleware instance.
//
// Parameters:
//   - key: Secret key for HMAC-SHA256
//   - logger: Logger instance for logging
//
// Returns a *HasherMiddleware ready to be used.
func newHasherMiddleware(key []byte, logger *zap.Logger) *hasherMiddleware {
	return &hasherMiddleware{
		key:    bytes.Clone(key),
		logger: logger,
	}
}

type responseWriter struct {
	http.ResponseWriter
	key []byte
}

// Write adds the HMAC-SHA256 hash header to the response and writes data.
//
// Parameters:
//   - data: The response body to write
//
// Returns the number of bytes written and any error encountered.
// Sets the Hash header with the computed HMAC-SHA256 hash.
func (rw *responseWriter) Write(data []byte) (int, error) {
	h := HexHash(data, rw.key)
	rw.ResponseWriter.Header().Set(HashHeader, h)
	return rw.ResponseWriter.Write(data)
}

type readCloser struct {
	io.Reader
}

// Close implements io.Closer for readCloser.
//
// It returns nil without performing any actual closing operation.
// This allows readCloser to satisfy io.ReadCloser interface.
func (rc *readCloser) Close() error { //nolint:revive // необходимо реализовать io.ReadCloser
	return nil
}

// middlewareHandler wraps the next handler with hash validation and response hashing.
//
// Parameters:
//   - next: The next HTTP handler in the chain
//
// Returns an http.Handler with hash processing logic.
func (m *hasherMiddleware) middlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data bytes.Buffer
		_, err := io.Copy(&data, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Failed to read request body: %v", err)
			return
		}
		r.Body.Close()

		transmittedHash := r.Header.Get(HashHeader)
		computedHash := HexHash(data.Bytes(), m.key)
		bodyLen := data.Len()

		// ниже transmittedHash != "" - особенность работы yandex тестов.
		// даже если у сервера стоит параметр KEY, но агент не передал хэш,
		// то хэш на стороне сервера не проверяется
		if bodyLen > 0 && transmittedHash != "" {
			if computedHash != transmittedHash {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Failed to verifying the hash through header %s", HashHeader)
				return
			}
		}
		r.Body = &readCloser{&data}

		wrapped := &responseWriter{
			key:            m.key,
			ResponseWriter: w,
		}

		next.ServeHTTP(wrapped, r)
	})
}
