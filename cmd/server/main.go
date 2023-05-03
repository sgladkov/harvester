package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"github.com/sgladkov/harvester/internal/utils"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var storage storage2.Storage

func main() {
	// check arguments
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	logLevel := flag.String("l", "info", "log level (fatal,  error,  warn, info, debug)")
	storeInterval := flag.Int("i", 300, "metrics store interval")
	fileStorage := flag.String("s", "/tmp/metrics-db.json", "file to store ans restore metrics")
	restoreFlag := flag.Bool("r", true, "should server read initial metrics value from the file")
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

	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if len(envStoreInterval) > 0 {
		val, err := strconv.ParseInt(envStoreInterval, 10, 32)
		if err != nil {
			logger.Log.Fatal("Failed to interpret STORE_INTERVAL environment variable",
				zap.String("value", envStoreInterval), zap.Error(err))
		}
		*storeInterval = int(val)
	}
	envFileStorage := os.Getenv("FILE_STORAGE_PATH")
	if len(envFileStorage) > 0 {
		*fileStorage = envFileStorage
	}
	envRestore, exist := os.LookupEnv("RESTORE")
	if exist {
		val, err := strconv.ParseBool(envRestore)
		if err != nil {
			logger.Log.Fatal("Failed to interpret RESTORE environment variable",
				zap.String("value", envRestore), zap.Error(err))
		}
		*restoreFlag = val
	}

	saveSettingsOnChange := *storeInterval == 0
	storage = storage2.NewMemStorage(*fileStorage, saveSettingsOnChange)
	if *restoreFlag {
		err := storage.Read()
		if err != nil {
			logger.Log.Warn("failed to read initial metrics values from file", zap.Error(err))
		}
	}

	if *storeInterval > 0 {
		storeTicker := time.NewTicker(time.Duration(*storeInterval) * time.Second)
		defer storeTicker.Stop()
		go func() {
			for range storeTicker.C {
				err := storage.Save()
				if err != nil {
					logger.Log.Warn("Failed to save metrics", zap.Error(err))
				}
				logger.Log.Info("Metrics are read")
			}
		}()
	}

	logger.Log.Info("Starting server", zap.String("address", *endpoint))
	err = http.ListenAndServe(*endpoint, MetricsRouter())
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
	err = storage.Save()
	if err != nil {
		logger.Log.Fatal("failed to store metrics", zap.Error(err))
	}
}

func MetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(logger.RequestLogger)
	r.Use(utils.GzipHandle)
	r.Get("/", getAllMetrics)
	r.Route("/update/", func(r chi.Router) {
		r.Post("/", updateMetricJSON)
		r.Post("/{type}/{name}/{value}", updateMetric)
	})
	r.Route("/value/", func(r chi.Router) {
		r.Post("/", getMetricJSON)
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
		http.Error(w, fmt.Sprintf("unknown metrics type [%s]", metricType), http.StatusBadRequest)
	}
}

func updateGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update gauge [%s]", err), http.StatusBadRequest)
		return
	}
	logger.Log.Debug("update gauge metric", zap.String("name", name), zap.Float64("value", value))
	err = storage.SetGauge(name, value)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update gauge [%s]", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
	if err != nil {
		logger.Log.Warn("failed to update counter", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update counter [%s]", err), http.StatusBadRequest)
		return
	}
	logger.Log.Debug("update counter metric", zap.String("name", name), zap.Int64("value", value))
	err = storage.SetCounter(name, value)
	if err != nil {
		logger.Log.Warn("failed to update counter", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update counter [%s]", err), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateMetricJSON(w http.ResponseWriter, r *http.Request) {
	if !utils.ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var m interfaces.Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Metrics", zap.Error(err))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", err), http.StatusBadRequest)
		return
	}
	logger.Log.Info("updateMetricJSON", zap.Any("metrics", m))
	m, err = storage.SetMetrics(m)
	if err != nil {
		logger.Log.Warn("Failed to save Metrics to storage", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to save Metrics to storage [%s]", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&m)
	if err != nil {
		logger.Log.Warn("Failed to write Metrics JSON to body", zap.Error(err))
		return
	}
}

func getAllMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
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
		http.Error(w, fmt.Sprintf("unknown metrics type [%s]", metricType), http.StatusBadRequest)
	}
}

func getGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := storage.GetGauge(name)
	if err != nil {
		logger.Log.Warn("failed to get gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get gauge [%s]", err), http.StatusNotFound)
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
		http.Error(w, fmt.Sprintf("failed to get counter [%s]", err), http.StatusNotFound)
		return
	}
	logger.Log.Debug("requested counter metric", zap.String("name", name), zap.Int64("value", value))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%d", value)))
	if err != nil {
		logger.Log.Warn("failed to write response body", zap.Error(err))
	}
}

func getMetricJSON(w http.ResponseWriter, r *http.Request) {
	if !utils.ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var m interfaces.Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Metrics")
		http.Error(w, fmt.Sprintf("Failed to decode JSON to Metrics [%s]", err), http.StatusBadRequest)
		return
	}
	//logger.Log.Info("getMetricJSON", zap.Any("metrics", m))
	m, err = storage.GetMetrics(m)
	if err != nil {
		logger.Log.Warn("Failed to get Metrics from storage")
		http.Error(w, fmt.Sprintf("Failed to get Metrics from storage [%s]", err), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&m)
	if err != nil {
		logger.Log.Warn("Failed to write Metrics JSON to body")
		return
	}
}
