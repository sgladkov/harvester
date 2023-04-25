package utils

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ContainsHeaderValue(r, "Accept-Encoding", "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func ContainsHeaderValue(r *http.Request, header string, value string) bool {
	values := r.Header.Values(header)
	for _, v := range values {
		fields := strings.FieldsFunc(v, func(c rune) bool { return c == ' ' || c == ',' || c == ';' })
		for _, f := range fields {
			if f == value {
				return true
			}
		}
	}
	return false
}
