// Package compress provides GZIP compression and decompression utilities.
//
// It offers functions for compressing and decompressing io.Reader streams
// using GZIP compression, as well as HTTP middleware handlers for automatic
// compression/decompression of HTTP request/response bodies.
package compress

import (
	"compress/gzip"
	"fmt"
	"io"
)

// CompressWithGZIP creates a reader that provides GZIP-compressed data.
//
// Parameters:
//   - r: Source data reader
//
// It reads data from the input reader, compresses it using GZIP, and returns
// a new reader that yields the compressed data.
//
// The compression happens asynchronously in a goroutine. The caller must close
// the returned reader when done.
//
// Returns an io.Reader that provides GZIP-compressed data.
func CompressWithGZIP(r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close() //nolint:errcheck // ignore close error

		gw := gzip.NewWriter(pw)
		defer gw.Close() //nolint:errcheck // ignore close error

		_, err := io.Copy(gw, r)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to copy: %w", err))
			return
		}
	}()
	return pr
}

// DecompressWithGZIP creates a reader that provides GZIP-decompressed data.
//
// Parameters:
//   - r: GZIP-compressed data reader
//
// It reads GZIP-compressed data from the input reader, decompresses it, and
// returns a new reader that yields the original uncompressed data.
//
// The decompression happens asynchronously in a goroutine. The caller must close
// the returned reader when done.
//
// Returns an io.Reader that provides decompressed data.
// Returns an error if the input stream is not valid GZIP data.
func DecompressWithGZIP(r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close() //nolint:errcheck // ignore close error

		gr, err := gzip.NewReader(r)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create gzip reader: %w", err))
			return
		}
		defer gr.Close() //nolint:errcheck // ignore close error

		_, err = io.Copy(pw, io.LimitReader(gr, 100<<20)) // Limit decompression to 100MB to prevent DoS
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to copy: %w", err))
			return
		}
	}()
	return pr
}
