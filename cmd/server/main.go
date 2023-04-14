package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"net/http"
	"os"
	"strconv"
)

var storage storage2.Storage

func main() {
	storage = storage2.NewMemStorage()
	// check arguments
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	flag.Parse()
	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*endpoint = address
	}

	fmt.Println(*endpoint)
	err := http.ListenAndServe(*endpoint, MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func MetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
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
		fmt.Printf("Unkvown metric type [%s]\n", metricType)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func updateGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Gauge metric: %s = %g\n", name, value)
	err = storage.SetGauge(name, value)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(storage)
	w.WriteHeader(http.StatusOK)
}

func updateCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Counter metric: %s = %d\n", name, value)
	err = storage.SetCounter(name, value)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(storage)
	w.WriteHeader(http.StatusOK)
}

func getAllMetrics(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(storage.GetAll()))
	if err != nil {
		fmt.Printf("Failed to write responce body: %s\n", err)
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
		fmt.Printf("Unkvown metric type [%s]\n", metricType)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := storage.GetGauge(name)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Printf("Gauge metric: %s = %g\n", name, value)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%g", value)))
	if err != nil {
		fmt.Printf("Failed to write responce body: %s\n", err)
	}
}

func getCounter(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := storage.GetCounter(name)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Printf("Counter metric: %s = %d\n", name, value)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%d", value)))
	if err != nil {
		fmt.Printf("Failed to write responce body: %s\n", err)
	}
}
