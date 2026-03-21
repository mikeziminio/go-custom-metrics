// Package compress provides GZIP compression and decompression HTTP middleware.
//
// It includes middleware handlers for automatically compressing outgoing
// responses and decompressing incoming request bodies based on HTTP headers.
package compress

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// responseWriter wraps an http.ResponseWriter to enable GZIP compression
// of written response data.
//
// It intercepts WriteHeader and Write calls to compress response data
// before sending it to the client. Only writes the first WriteHeader call.
type responseWriter struct {
	http.ResponseWriter
	defaultStatusCode int
	written           bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

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
// It checks the Accept-Encoding header for "gzip" support, and if present,
// wraps the response writer to compress all response data before sending.
//
// If the client doesn't support gzip compression, the middleware passes
// the request through without modification.
//
// Parameter next is the next HTTP handler in the chain.
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

// readCloser wraps an io.Reader to satisfy io.ReadCloser interface.
//
// The Close method is implemented as a no-op since the underlying
// reader doesn't need to be closed.
type readCloser struct {
	io.Reader
}

// Close is a no-op implementation of io.Closer for readCloser.
//
// Since the underlying reader doesn't need to be closed, this method
// simply returns nil without performing any action.
func (rc *readCloser) Close() error { //nolint:revive // необходимо реализовать io.ReadCloser
	return nil
}

// DecompressMiddlewareHandler creates an HTTP middleware that decompresses
// request bodies using GZIP when the body is compressed.
//
// It checks the Content-Encoding header for "gzip" value, and if present,
// decompresses the request body before passing it to the next handler.
//
// Parameter next is the next HTTP handler in the chain.
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
