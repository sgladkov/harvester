package httprouter

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	//_ "github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/models"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

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
	err := models.ValidateMetricsID(name)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update gauge [%s]", err), http.StatusBadRequest)
		return
	}
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
	err := models.ValidateMetricsID(name)
	if err != nil {
		logger.Log.Warn("failed to update gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update gauge [%s]", err), http.StatusBadRequest)
		return
	}
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
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var m models.Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Metrics", zap.Error(err))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", err), http.StatusBadRequest)
		return
	}
	err = models.ValidateMetricsID(m.ID)
	if err != nil {
		logger.Log.Warn("failed to update metrics", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to update metrics [%s]", err), http.StatusBadRequest)
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
	err := models.ValidateMetricsID(name)
	if err != nil {
		logger.Log.Warn("failed to get gauge", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get gauge [%s]", err), http.StatusBadRequest)
		return
	}
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
	err := models.ValidateMetricsID(name)
	if err != nil {
		logger.Log.Warn("failed to get counter", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get counter [%s]", err), http.StatusBadRequest)
		return
	}
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
	if !ContainsHeaderValue(r, "Content-Type", "application/json") {
		contentType := r.Header.Get("Content-Type")
		logger.Log.Warn("Wrong Content-Type header", zap.String("Content-Type", contentType))
		http.Error(w, fmt.Sprintf("Wrong Content-Type header [%s]", contentType), http.StatusBadRequest)
		return
	}
	var m models.Metrics
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		logger.Log.Warn("Failed to decode JSON to Metrics")
		http.Error(w, fmt.Sprintf("Failed to decode JSON to Metrics [%s]", err), http.StatusBadRequest)
		return
	}
	err = models.ValidateMetricsID(m.ID)
	if err != nil {
		logger.Log.Warn("failed to get metrics", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get metrics [%s]", err), http.StatusBadRequest)
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

func ping(w http.ResponseWriter, _ *http.Request) {
	logger.Log.Info("ping", zap.String("DSN", databaseDSN))
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		logger.Log.Warn("Failed to connect DB", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Log.Warn("failed to close database", zap.Error(err))
		}
	}()
	w.WriteHeader(http.StatusOK)
}
