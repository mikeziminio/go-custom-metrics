package hasher

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewHasherMiddleware(t *testing.T) {
	key := []byte("test-key")

	middleware := newHasherMiddleware(key, zap.NewNop())
	assert.NotNil(t, middleware)
	assert.Equal(t, key, middleware.key)
}

func TestHasherMiddleware_MiddlewareHandler(t *testing.T) {
	key := []byte("test-key")

	t.Run("should accept valid request with correct hash", func(t *testing.T) {
		data := []byte("test data")
		hash := HexHash(data, key)

		req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
		req.Header.Set(HashHeader, hash)

		rr := httptest.NewRecorder()

		middleware := newHasherMiddleware(key, zap.NewNop())
		handler := middleware.middlewareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, data, body)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should reject request with invalid hash", func(t *testing.T) {
		data := []byte("test data")
		invalidHash := "invalid-hash"

		req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
		req.Header.Set(HashHeader, invalidHash)

		rr := httptest.NewRecorder()

		middleware := newHasherMiddleware(key, zap.NewNop())
		handler := middleware.middlewareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should set hash header in response", func(t *testing.T) {
		data := []byte("response data")
		req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
		req.Header.Set(HashHeader, HexHash(data, key))

		rr := httptest.NewRecorder()

		middleware := newHasherMiddleware(key, zap.NewNop())
		handler := middleware.middlewareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get(HashHeader))
	})

	t.Run("should handle empty request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte{}))
		req.Header.Set(HashHeader, HexHash([]byte{}, key))

		rr := httptest.NewRecorder()

		middleware := newHasherMiddleware(key, zap.NewNop())
		handler := middleware.middlewareHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Empty(t, body)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestResponseWriter_Write(t *testing.T) {
	key := []byte("test-key")

	t.Run("should set hash header in response", func(t *testing.T) {
		rr := httptest.NewRecorder()
		responseWriter := &responseWriter{
			key:            key,
			ResponseWriter: rr,
		}

		data := []byte("test response data")
		n, err := responseWriter.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.NotEmpty(t, rr.Header().Get(HashHeader))
	})
}

func TestReadCloser_Close(t *testing.T) {
	rc := &readCloser{}
	err := rc.Close()
	assert.NoError(t, err)
}
