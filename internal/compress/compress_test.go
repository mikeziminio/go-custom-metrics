package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressWithGZIP(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "small data",
			input: "some small data",
		},
		{
			name:  "large data",
			input: strings.Repeat("some large data ", 1000),
		},
		{
			name:  "empty data",
			input: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			compressedReader := CompressWithGZIP(strings.NewReader(tc.input))

			var buf bytes.Buffer
			_, err := io.Copy(&buf, compressedReader)
			require.NoError(t, err)

			// Декомпрессируем данные для проверки их корректности
			reader2 := bytes.NewReader(buf.Bytes())
			gzipReader, err := gzip.NewReader(reader2)
			require.NoError(t, err)
			defer gzipReader.Close() //nolint:errcheck // ignore close error

			decompressedBytes, err := io.ReadAll(gzipReader)
			require.NoError(t, err)

			// Проверяем, что декомпрессированные данные соответствуют оригиналу
			assert.Equal(t, tc.input, string(decompressedBytes))
		})
	}
}

func TestDecompressWithGZIP(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "small data",
			input: "some small data",
		},
		{
			name:  "large data",
			input: strings.Repeat("some large data ", 1000),
		},
		{
			name:  "empty data",
			input: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Сначала сжимаем данные с помощью стандартной библиотеки
			var buf bytes.Buffer
			writer := gzip.NewWriter(&buf)
			_, err := writer.Write([]byte(tc.input))
			require.NoError(t, err)
			err = writer.Close()
			require.NoError(t, err)

			// Теперь декомпрессируем с помощью нашей функции
			decompressedReader := DecompressWithGZIP(bytes.NewReader(buf.Bytes()))

			decompressedBytes, err := io.ReadAll(decompressedReader)
			require.NoError(t, err)

			// Проверяем, что декомпрессированные данные соответствуют оригиналу
			assert.Equal(t, tc.input, string(decompressedBytes))
		})
	}
}
