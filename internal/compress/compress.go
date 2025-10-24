package compress

import (
	"compress/gzip"
	"fmt"
	"io"
)

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

		_, err = io.Copy(pw, gr)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to copy: %w", err))
			return
		}
	}()
	return pr
}
