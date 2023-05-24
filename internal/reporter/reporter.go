package reporter

import (
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/models"
	"math/rand"
	"runtime"
	"sync"
)

type Reporter struct {
	connection interfaces.ServerConnection
	gauges     map[string]float64
	counters   map[string]int64
	lock       sync.Mutex
}

func NewReporter(connection interfaces.ServerConnection) *Reporter {
	result := Reporter{
		connection: connection,
		gauges:     make(map[string]float64),
		counters:   make(map[string]int64),
	}
	result.counters["PollCount"] = 0
	return &result
}

func (m *Reporter) Poll() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	data := runtime.MemStats{}
	runtime.ReadMemStats(&data)
	m.gauges["Alloc"] = float64(data.Alloc)
	m.gauges["BuckHashSys"] = float64(data.BuckHashSys)
	m.gauges["Frees"] = float64(data.Frees)
	m.gauges["FreeMemory"] = float64(data.Frees)
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
	m.gauges["TotalMemory"] = float64(data.TotalAlloc)
	m.gauges["CPUutilization1"] = float64(1)
	m.gauges["RandomValue"] = rand.Float64()
	m.counters["PollCount"]++
	return nil
}

func (m *Reporter) Report() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	metrics := models.Metrics{}
	for name, value := range m.gauges {
		metrics.MType = "gauge"
		metrics.ID = name
		metrics.Value = &value
		err := m.connection.UpdateMetrics(&metrics)
		if err != nil {
			return err
		}
	}
	for name, value := range m.counters {
		metrics.MType = "counter"
		metrics.ID = name
		metrics.Delta = &value
		err := m.connection.UpdateMetrics(&metrics)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Reporter) BatchReport() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	var batch []models.Metrics
	for name, value := range m.gauges {
		metrics := models.Metrics{}
		metrics.MType = "gauge"
		metrics.ID = name
		metrics.Value = &value
		batch = append(batch, metrics)
	}
	for name, value := range m.counters {
		metrics := models.Metrics{}
		metrics.MType = "counter"
		metrics.ID = name
		metrics.Delta = &value
		batch = append(batch, metrics)
	}
	return m.connection.BatchUpdateMetrics(batch)
}
