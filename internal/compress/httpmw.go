package compress

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

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

type readCloser struct {
	io.Reader
}

func (rc *readCloser) Close() error {
	return nil
}

func DecompressMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			decompressedReader := DecompressWithGZIP(r.Body)
			r.Body = &readCloser{
				Reader: decompressedReader,
			}
		}
		next.ServeHTTP(w, r)
	})
}
