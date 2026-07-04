package compress

import (
	"compress/gzip"
	"net/http"
	"slices"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	contentTypeForCompress := slices.Contains(TypesForGzip, c.Header().Get("Content-Type"))
	if contentTypeForCompress {
		return c.zw.Write(p)
	}

	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		contentTypeForCompress := slices.Contains(TypesForGzip, c.Header().Get("Content-Type"))
		if contentTypeForCompress {
			c.w.Header().Set("Content-Encoding", "gzip")
			c.w.Header().Del("Content-Length")
		}
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}
