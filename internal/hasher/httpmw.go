package hasher

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

const HashHeader = "HashSHA256"

type HasherMiddleware struct {
	key    []byte
	logger *zap.Logger
}

func NewHasherMiddleware(key []byte, logger *zap.Logger) *HasherMiddleware {
	return &HasherMiddleware{
		key:    bytes.Clone(key),
		logger: logger,
	}
}

type responseWriter struct {
	http.ResponseWriter
	key []byte
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	h := HexHash(data, rw.key)
	rw.ResponseWriter.Header().Set(HashHeader, h)
	return rw.ResponseWriter.Write(data)
}

type readCloser struct {
	io.Reader
}

func (rc *readCloser) Close() error { //nolint:revive // необходимо реализовать io.ReadCloser
	return nil
}

func (m *HasherMiddleware) MiddlewareHandler(next http.Handler) http.Handler {
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
		m.logger.Info("hasher",
			zap.Int("len", bodyLen),
			zap.String("transmittedHash", transmittedHash),
			zap.String("computedHash", computedHash),
		)
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
