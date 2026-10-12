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

func (w *HashResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *HashResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}
