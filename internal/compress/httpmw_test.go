package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressMiddlewareHandler(t *testing.T) {
	testCases := []struct {
		name           string
		headers        http.Header
		expectBodyFunc func(*testing.T, []byte) []byte
	}{
		{
			name: "accept encoding with gzip",
			headers: map[string][]string{
				"Accept-Encoding": {"gzip"},
			},
			expectBodyFunc: decompressedBytes,
		},
		{
			name: "accept encoding with gzip and other encodings",
			headers: map[string][]string{
				"Accept-Encoding": {"deflate, gzip;q=1.0, *;q=0.5"},
			},
			expectBodyFunc: decompressedBytes,
		},
		{
			name: "accept encoding with wildcard",
			headers: map[string][]string{
				"Accept-Encoding": {"*"},
			},
			expectBodyFunc: decompressedBytes,
		},
		{
			name:           "no accept encoding header",
			headers:        map[string][]string{},
			expectBodyFunc: originalBytes,
		},
		{
			name: "accept encoding without gzip",
			headers: map[string][]string{
				"Accept-Encoding": {"deflate"},
			},
			expectBodyFunc: originalBytes,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalBody := []byte("some body data")

			// простой обработчик обёрнутый в мидлварю
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, err := w.Write(originalBody)
				assert.NoError(t, err)
			})
			wrappedHandler := CompressMiddlewareHandler(handler)

			// Запрос с заголовками, которые должны либо привести к тому,
			// что тело ответа закодируется, либо нет
			req := httptest.NewRequest("GET", "/", http.NoBody)
			req.Header = tc.headers

			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)

			body, err := io.ReadAll(rec.Body)
			require.NoError(t, err)

			// проверяем тело ответа
			assert.Equal(t,
				originalBody,
				tc.expectBodyFunc(t, body),
			)
		})
	}
}

func TestDecompressMiddlewareHandler(t *testing.T) {
	testCases := []struct {
		name            string
		headers         http.Header
		requestBodyFunc func(*testing.T, []byte) []byte
	}{
		{
			name: "content encoding gzip",
			headers: map[string][]string{
				"Content-Encoding": {"gzip"},
			},
			requestBodyFunc: compressedBytes,
		},
		{
			name:            "no content encoding header",
			headers:         map[string][]string{},
			requestBodyFunc: originalBytes,
		},
		{
			name: "content encoding other",
			headers: map[string][]string{
				"Content-Encoding": {"deflate"},
			},
			requestBodyFunc: originalBytes,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalBody := []byte("some body data")

			// запрос, в котором либо закодированное,
			// либо оригинальное тело
			req := httptest.NewRequest("POST", "/", bytes.NewReader(
				tc.requestBodyFunc(t, originalBody),
			))
			req.Header = tc.headers

			// простой обработчик, который ничего не делает
			// просто сохраняем тело запроса
			var body []byte
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				body, err = io.ReadAll(r.Body)
				assert.NoError(t, err)
				w.WriteHeader(http.StatusOK)
			})
			wrappedHandler := DecompressMiddlewareHandler(handler)
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			// проверяем, что тело запроса было декомпрессировано
			assert.Equal(t, originalBody, body)
		})
	}
}

func compressedBytes(t *testing.T, original []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(original)
	require.NoError(t, err)
	err = writer.Flush()
	require.NoError(t, err)
	_ = writer.Close() // write gzip footer
	compressed, err := io.ReadAll(&buf)
	require.NoError(t, err)
	return compressed
}

func decompressedBytes(t *testing.T, original []byte) []byte {
	t.Helper()
	reader := bytes.NewReader(original)
	gzipReader, err := gzip.NewReader(reader)
	require.NoError(t, err)
	defer gzipReader.Close() //nolint:errcheck // ignore close error
	decompressed, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	return decompressed
}

func originalBytes(t *testing.T, original []byte) []byte {
	t.Helper()
	return original
}
