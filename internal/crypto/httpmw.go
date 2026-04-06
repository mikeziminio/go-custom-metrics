package crypto

import (
	"bytes"
	"io"
	"net/http"

	"crypto/rsa"

	"go.uber.org/zap"
)

// MiddlewareHandler creates an HTTP middleware that decrypts request bodies using RSA-OAEP.
func MiddlewareHandler(key *rsa.PrivateKey, logger *zap.Logger) func(http.Handler) http.Handler {
	if key == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			if contentEncoding != "rsa-oaep" {
				next.ServeHTTP(w, r)
				return
			}

			var body bytes.Buffer
			_, err := io.Copy(&body, r.Body)
			if err != nil {
				r.Body.Close()
				w.WriteHeader(http.StatusBadRequest)
				logger.Error("Failed to read request body", zap.Error(err))
				return
			}
			r.Body.Close()

			decrypted, err := DecryptWithPrivateKey(body.Bytes(), key)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Error("Failed to decrypt request body", zap.Error(err))
				return
			}

			r.Body = &readCloser{
				Reader: bytes.NewReader(decrypted),
			}

			next.ServeHTTP(w, r)
		})
	}
}

type readCloser struct {
	io.Reader
}

func (rc *readCloser) Close() error {
	return nil
}
