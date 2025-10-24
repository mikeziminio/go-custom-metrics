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

var testBody = []byte("test data for compression")

func TestCompressMiddlewareHandler(t *testing.T) {
	testCases := []struct {
		name           string
		acceptEncoding string
		expectCompress bool
	}{
		{
			name:           "no accept encoding header",
			acceptEncoding: "",
			expectCompress: false,
		},
		{
			name:           "accept encoding with gzip",
			acceptEncoding: "gzip",
			expectCompress: true,
		},
		{
			name:           "accept encoding with gzip and other encodings",
			acceptEncoding: "deflate, gzip;q=1.0, *;q=0.5",
			expectCompress: true,
		},
		{
			name:           "accept encoding with wildcard",
			acceptEncoding: "*",
			expectCompress: true,
		},
		{
			name:           "accept encoding without gzip",
			acceptEncoding: "deflate",
			expectCompress: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// простой обработчик обёрнутый в мидлварю
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, err := w.Write(testBody)
				assert.NoError(t, err)
			})
			wrappedHandler := CompressMiddlewareHandler(handler)

			req := httptest.NewRequest("GET", "/", http.NoBody)
			if tc.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			}
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			if tc.expectCompress {
				reader := bytes.NewReader(rec.Body.Bytes())
				gzipReader, err := gzip.NewReader(reader)
				require.NoError(t, err)
				defer gzipReader.Close() //nolint:errcheck // ignore close error

				decompressedData, err := io.ReadAll(gzipReader)
				require.NoError(t, err)
				assert.Equal(t, testBody, decompressedData)
			} else {
				assert.Equal(t, testBody, rec.Body.Bytes())
			}
		})
	}
}

func TestDecompressMiddlewareHandler(t *testing.T) {
	testCases := []struct {
		name             string
		contentEncoding  string
		requestBody      []byte
		shouldDecompress bool
	}{
		{
			name:             "no content encoding header",
			contentEncoding:  "",
			requestBody:      []byte("test data"),
			shouldDecompress: false,
		},
		{
			name:             "content encoding gzip",
			contentEncoding:  "gzip",
			requestBody:      []byte("test data"),
			shouldDecompress: true,
		},
		{
			name:             "content encoding other",
			contentEncoding:  "deflate",
			requestBody:      []byte("test data"),
			shouldDecompress: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with appropriate body
			var req *http.Request
			if tc.contentEncoding == "gzip" {
				// Create gzipped request body
				var buf bytes.Buffer
				writer := gzip.NewWriter(&buf)
				_, err := writer.Write(tc.requestBody)
				require.NoError(t, err)
				err = writer.Close()
				require.NoError(t, err)

				req = httptest.NewRequest("POST", "/", &buf)
				req.Header.Set("Content-Encoding", "gzip")
			} else {
				req = httptest.NewRequest("POST", "/", bytes.NewReader(tc.requestBody))
				if tc.contentEncoding != "" {
					req.Header.Set("Content-Encoding", tc.contentEncoding)
				}
			}

			// Simple handler that reads the body and returns it
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				w.Header().Set("Content-Type", "text/plain")
				_, err = w.Write(body)
				assert.NoError(t, err)
			})

			wrappedHandler := DecompressMiddlewareHandler(handler)
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tc.requestBody, rec.Body.Bytes())
		})
	}
}
