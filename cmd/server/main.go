package main

import (
	"errors"
	"fmt"
	"github.com/sgladkov/harvester/internal"
	"net/http"
	"strconv"
	"strings"
)

var storage internal.Storage

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage = internal.NewMemStorage()
	http.Handle("/update/", http.StripPrefix("/update/", http.HandlerFunc(webhook)))
	return http.ListenAndServe(`:8080`, nil)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Url:", r.URL)
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	status, err := ProcessMetric(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
}

func ProcessMetric(data string) (int, error) {
	if len(data) == 0 {
		return http.StatusNotFound, errors.New("empty metrica data")
	}
	parts := strings.Split(data, "/")
	if len(parts) != 3 {
		return http.StatusNotFound, errors.New("invalid metric data format")
	}
	var name string
	switch parts[0] {
	case "gauge":
		name = parts[1]
		value, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			fmt.Println(err)
			return http.StatusBadRequest, errors.New("wrong gauge metric value format")
		}
		fmt.Printf("Gauge metric: %s = %f\n", name, value)
		storage.SetGauge(name, value)
		fmt.Println(storage)
	case "counter":
		name = parts[1]
		value, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			fmt.Println(err)
			return http.StatusBadRequest, errors.New("wrong counter metric value format")
		}
		fmt.Printf("Counter metric: %s = %d\n", name, value)
		storage.SetCounter(name, value)
		fmt.Println(storage)
	default:
		return http.StatusBadRequest, fmt.Errorf("unknown metric type [%s]", parts[0])
	}
	return http.StatusOK, nil
}
