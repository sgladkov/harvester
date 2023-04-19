package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sgladkov/harvester/internal/logger"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strconv"
)

var storage storage2.Storage

func main() {
	storage = storage2.NewMemStorage()
	// check arguments
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	logLevel := flag.String("l", "info", "log level (fatal,  error,  warn, info, debug)")
	flag.Parse()
	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*endpoint = address
	}
	envLogLevel := os.Getenv("LOG_LEVEL")
	if len(envLogLevel) > 0 {
		*logLevel = envLogLevel
	}

	err := logger.Initialize(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Starting server", zap.String("address", *endpoint))
	err = http.ListenAndServe(*endpoint, MetricsRouter())
	if err != nil {
		log.Fatal(err)
	}
}

func MetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(logger.RequestLogger)
	r.Get("/", getAllMetrics)
	r.Route("/update/", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", updateMetric)
	})
	r.Route("/value/", func(r chi.Router) {
		r.Get("/{type}/{name}", getMetric)
	})
	return r
}

func updateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	switch metricType {
	case "gauge":
		updateGauge(w, r)
	case "counter":
		updateCounter(w, r)
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", metricType))
		w.WriteHeader(http.StatusBadRequest)
	}
}

func updateGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Debug("update gauge metric", zap.String("name", name), zap.Float64("value", value))
	err = storage.SetGauge(name, value)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
	if err != nil {
		logger.Log.Warn("failed to update counter", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Debug("update counter metric", zap.String("name", name), zap.Int64("value", value))
	err = storage.SetCounter(name, value)
	if err != nil {
		logger.Log.Warn("failed to update counter", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getAllMetrics(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(storage.GetAll()))
	if err != nil {
		logger.Log.Warn("failed to write response body", zap.Error(err))
	}
}

func getMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	switch metricType {
	case "gauge":
		getGauge(w, r)
	case "counter":
		getCounter(w, r)
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", metricType))
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := storage.GetGauge(name)
	if err != nil {
		logger.Log.Warn("failed to get gauge", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logger.Log.Debug("requested gauge metric", zap.String("name", name), zap.Float64("value", value))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%g", value)))
	if err != nil {
		logger.Log.Warn("failed to write response body", zap.Error(err))
	}
}

func getCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := storage.GetCounter(name)
	if err != nil {
		logger.Log.Warn("failed to get counter", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logger.Log.Debug("requested counter metric", zap.String("name", name), zap.Int64("value", value))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%d", value)))
	if err != nil {
		logger.Log.Warn("failed to write response body", zap.Error(err))
	}
}
