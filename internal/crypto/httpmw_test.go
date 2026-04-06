package crypto

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddlewareHandler_Success(t *testing.T) {
	privKey, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)

	logger := zap.NewNop()

	originalData := map[string]string{"test": "data"}
	originalBody, err := json.Marshal(originalData)
	assert.NoError(t, err)

	pubKey, err := LoadPublicKey(testPubKeyPath)
	assert.NoError(t, err)

	encryptedBody, err := EncryptWithPublicKey(originalBody, pubKey)
	assert.NoError(t, err)

	var handlerBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		handlerBody = body
		w.WriteHeader(http.StatusOK)
	})

	middleware := MiddlewareHandler(privKey, logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(encryptedBody))
	req.Header.Set("Content-Encoding", "rsa-oaep")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, string(originalBody), string(handlerBody))
}

func TestMiddlewareHandler_NoEncoding(t *testing.T) {
	privKey, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)

	logger := zap.NewNop()

	handlerCallCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCallCount++
		w.WriteHeader(http.StatusOK)
	})

	middleware := MiddlewareHandler(privKey, logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/", http.NoBody)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, handlerCallCount)
}

func TestMiddlewareHandler_InvalidKey(t *testing.T) {
	logger := zap.NewNop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := MiddlewareHandler(nil, logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/", http.NoBody)
	req.Header.Set("Content-Encoding", "rsa-oaep")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddlewareHandler_DecryptError(t *testing.T) {
	privKey, err := LoadPrivateKey(testPrivKeyPath)
	assert.NoError(t, err)

	logger := zap.NewNop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := MiddlewareHandler(privKey, logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("invalid encrypted data")))
	req.Header.Set("Content-Encoding", "rsa-oaep")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
