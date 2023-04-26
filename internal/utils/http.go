package utils

import (
	"compress/gzip"
	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
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
		writerToUse := w
		if ContainsHeaderValue(r, "Content-Encoding", "gzip") {
			// change original request body to decode its content
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			logger.Log.Info("Use request decode")
			r.Body = gz
		} else {
			logger.Log.Info("Don't use request decode")
		}

		if ContainsHeaderValue(r, "Accept-Encoding", "gzip") {
			// change writer to wrapped writer with gzip encoding
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer func() {
				err = gz.Close()
				if err != nil {
					logger.Log.Warn("Failed to close gzip writer", zap.Error(err))
				}
			}()

			logger.Log.Info("Use reply compress")
			w.Header().Set("Content-Encoding", "gzip")
			writerToUse = gzipWriter{ResponseWriter: w, Writer: gz}
		} else {
			// use original writer
			logger.Log.Info("Don't use reply compress")
		}

		h.ServeHTTP(writerToUse, r)
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
