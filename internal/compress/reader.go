package compress

import (
	"compress/gzip"
	"io"
)

type compressRequestReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressRequestReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressRequestReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressRequestReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressRequestReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
