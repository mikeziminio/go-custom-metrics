package compress

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// responseWriter wraps an http.responseWriter to compress response data
// using GZIP compression.
//
// It implements http.responseWriter and automatically sets the
// Content-Encoding header to "gzip" before writing the first response.
//
// The WriteHeader method is idempotent - it only writes the header once,
// ensuring the Content-Encoding header is set before any data is written.
type responseWriter struct {
	http.ResponseWriter
	defaultStatusCode int
	written           bool
}

// WriteHeader writes the HTTP response header with gzip encoding set.
//
// Parameters:
//   - code: The HTTP status code to write
//
// This method is idempotent - it only writes the header once and sets
// the Content-Encoding header to "gzip".
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

// Write compresses and writes data to the response.
//
// Parameters:
//   - data: The byte slice to compress and write
//
// Returns the number of bytes written and any error encountered.
// Automatically calls WriteHeader with the default status code if not
// already called.
func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(rw.defaultStatusCode)
	}
	r := CompressWithGZIP(bytes.NewReader(data))
	n, err := io.Copy(rw.ResponseWriter, r)
	return int(n), err
}

// CompressMiddlewareHandler creates an HTTP middleware that compresses
// response data using GZIP when the client supports it.
//
// Parameters:
//   - next: The next HTTP handler in the chain
//
// It checks the Accept-Encoding header for "gzip" support, and if present,
// wraps the response writer to compress all response data before sending.
//
// If the client doesn't support gzip compression, the middleware passes
// the request through without modification.
//
// Returns an http.Handler that compresses responses when supported.
func CompressMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ae := r.Header.Get("Accept-Encoding")
		// проверка не самая строгая, для продакшна нужно улучшить
		if ae != "*" && !strings.Contains(ae, "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		wrapped := &responseWriter{
			ResponseWriter:    w,
			defaultStatusCode: http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
	})
}

// readCloser wraps an io.Reader to implement io.readCloser.
//
// It provides a Close method that does nothing, fulfilling the
// io.readCloser interface requirement without actual cleanup.
type readCloser struct {
	io.Reader
}

// Close implements io.Closer for ReadCloser.
//
// It returns nil without performing any actual closing operation.
// This allows ReadCloser to satisfy io.ReadCloser interface.
func (rc *readCloser) Close() error { //nolint:revive // необходимо реализовать io.ReadCloser
	return nil
}

// DecompressMiddlewareHandler creates an HTTP middleware that decompresses
// request bodies using GZIP when the body is compressed.
//
// It checks the Content-Encoding header for "gzip" value, and if present,
// decompresses the request body before passing it to the next handler.
//
// Parameters:
//   - next: The next HTTP handler in the chain
//
// Returns an http.Handler that decompresses request bodies when compressed.
func DecompressMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			decompressedReader := DecompressWithGZIP(r.Body)
			defer r.Body.Close()
			r.Body = &readCloser{
				Reader: decompressedReader,
			}
		}
		next.ServeHTTP(w, r)
	})
}
