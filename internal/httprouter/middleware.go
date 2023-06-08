package httprouter

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"time"

	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
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

type (
	responseData struct {
		status int
		size   int
		body   []byte
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.body = append(r.responseData.body, b...)
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		logger.Log.Info("request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
}

func HandleHash(h http.Handler, key []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			h.ServeHTTP(w, r)
			return
		}
		if r.URL.Path != "updates/" {
			h.ServeHTTP(w, r)
			return
		}
		msgHash := r.Header.Get("HashSHA256")
		if len(msgHash) == 0 {
			//http.Error(w, "no data signature", http.StatusBadRequest)
			h.ServeHTTP(w, r)
			return
		}
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Warn("error reading body", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		hash, err := HashFromData(msg, key)
		if err != nil {
			logger.Log.Warn("failed to sign data", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if hash != msgHash {
			logger.Log.Warn("invalid signature")
			http.Error(w, "unvalid signature", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(msg))
		h.ServeHTTP(w, r)
	})
}
