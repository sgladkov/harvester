package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sgladkov/harvester/internal"
	"net/http"
	"strconv"
)

var storage internal.Storage

func main() {
	storage = internal.NewMemStorage()
	err := http.ListenAndServe(`:8080`, MetricsRouter())
	if err != nil {
		panic(err)
	}
}

func MetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/update/", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", updateGauge)
		r.Post("/counter/{name}/{value}", updateCounter)
	})
	return r
}

func updateGauge(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Gauge metric: %s = %f\n", name, value)
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
