package compress

import (
	"net/http"
	"strings"
)

var TypesForGzip = []string{"text/html", "application/json"}

func GzipHandler(h http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contentGzip := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
		if contentGzip {
			compressReader, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = compressReader
			defer compressReader.Close()
		}

		acceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if acceptGzip {
			cw := newCompressWriter(w)
			w = cw
			defer cw.Close()
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
		}

		h.ServeHTTP(w, r)
	}

	return fn
}
