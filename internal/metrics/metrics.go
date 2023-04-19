package metrics

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
)

type Metrics struct {
	server   string
	gauges   map[string]float64
	counters map[string]int64
	lock     sync.Mutex
}

func NewMetrics(server string) *Metrics {
	result := Metrics{}
	result.server = server
	result.gauges = make(map[string]float64)
	result.counters = make(map[string]int64)
	result.counters["PollCount"] = 0
	return &result
}

func (m *Metrics) Poll() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	data := runtime.MemStats{}
	runtime.ReadMemStats(&data)
	m.gauges["Alloc"] = float64(data.Alloc)
	m.gauges["BuckHashSys"] = float64(data.BuckHashSys)
	m.gauges["Frees"] = float64(data.Frees)
	m.gauges["GCCPUFraction"] = data.GCCPUFraction
	m.gauges["GCSys"] = float64(data.GCSys)
	m.gauges["HeapAlloc"] = float64(data.HeapAlloc)
	m.gauges["HeapIdle"] = float64(data.HeapIdle)
	m.gauges["HeapInuse"] = float64(data.HeapInuse)
	m.gauges["HeapObjects"] = float64(data.HeapObjects)
	m.gauges["HeapReleased"] = float64(data.HeapReleased)
	m.gauges["HeapSys"] = float64(data.HeapSys)
	m.gauges["LastGC"] = float64(data.LastGC)
	m.gauges["Lookups"] = float64(data.Lookups)
	m.gauges["MCacheInuse"] = float64(data.MCacheInuse)
	m.gauges["MCacheSys"] = float64(data.MCacheSys)
	m.gauges["MSpanInuse"] = float64(data.MSpanInuse)
	m.gauges["MSpanSys"] = float64(data.MSpanSys)
	m.gauges["Mallocs"] = float64(data.Mallocs)
	m.gauges["NextGC"] = float64(data.NextGC)
	m.gauges["NumForcedGC"] = float64(data.NumForcedGC)
	m.gauges["NumGC"] = float64(data.NumGC)
	m.gauges["OtherSys"] = float64(data.OtherSys)
	m.gauges["PauseTotalNs"] = float64(data.PauseTotalNs)
	m.gauges["StackInuse"] = float64(data.StackInuse)
	m.gauges["StackSys"] = float64(data.StackSys)
	m.gauges["Sys"] = float64(data.Sys)
	m.gauges["TotalAlloc"] = float64(data.TotalAlloc)
	m.gauges["RandomValue"] = rand.Float64()
	m.counters["PollCount"]++
	return nil
}

func processRequest(client *http.Client, request *http.Request) error {
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	_, err = io.ReadAll(resp.Body)
	return err
}

func (m *Metrics) Report() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	client := &http.Client{}
	for name, value := range m.gauges {
		request, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/update/gauge/%s/%f", m.server, name, value),
			nil)
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "text/plain")
		err = processRequest(client, request)
		if err != nil {
			return err
		}
	}
	for name, value := range m.counters {
		request, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/update/counter/%s/%d", m.server, name, value),
			nil)
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "text/plain")
		err = processRequest(client, request)
		if err != nil {
			return err
		}
	}
	return nil
}
