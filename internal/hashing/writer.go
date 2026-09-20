package hashing

import (
	"bytes"
	"net/http"
)

type HashResponseWriter struct {
	w http.ResponseWriter
	b *bytes.Buffer
}

func (w *HashResponseWriter) Write(p []byte) (int, error) {
	return w.b.Write(p)
}

func (c *HashResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *HashResponseWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}
