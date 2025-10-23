package compress

import (
	"compress/gzip"
	"fmt"
	"io"
)

func CompressWithGZIP(r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()

		gw := gzip.NewWriter(pw)
		defer gw.Close()

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
		defer pw.Close()

		gr, err := gzip.NewReader(r)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create gzip reader: %w", err))
			return
		}
		defer gr.Close()

		_, err = io.Copy(pw, gr)
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to copy: %w", err))
			return
		}
	}()
	return pr
}
