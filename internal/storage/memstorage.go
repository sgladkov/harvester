package storage

import (
	"errors"
	"fmt"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/metrics"
	"go.uber.org/zap"
	"sync"
)

type Storage interface {
	GetGauge(name string) (float64, error)
	SetGauge(name string, value float64) error
	GetCounter(name string) (int64, error)
	SetCounter(name string, value int64) error
	GetAll() string
	SetMetrics(m metrics.Metrics) (metrics.Metrics, error)
	GetMetrics(m metrics.Metrics) (metrics.Metrics, error)
}

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	lock     sync.Mutex
}

func NewMemStorage() *MemStorage {
	result := MemStorage{}
	result.gauges = make(map[string]float64)
	result.counters = make(map[string]int64)
	return &result
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, exists := m.gauges[name]
	if !exists {
		return 0.0, fmt.Errorf("no gauge [%s]", name)
	}
	return value, nil
}

func (m *MemStorage) SetGauge(name string, value float64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.gauges[name] = value
	return nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, exists := m.counters[name]
	if !exists {
		return 0, fmt.Errorf("no counter [%s]", name)
	}
	return value, nil
}

func (m *MemStorage) SetCounter(name string, value int64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.counters[name] += value
	return nil
}

func (m *MemStorage) GetAll() string {
	m.lock.Lock()
	defer m.lock.Unlock()
	var res string
	for n, v := range m.gauges {
		res += fmt.Sprintf("%s=%g\n", n, v)
	}
	for n, v := range m.counters {
		res += fmt.Sprintf("%s=%d\n", n, v)
	}
	return res
}

func (s *MemStorage) SetMetrics(m metrics.Metrics) (metrics.Metrics, error) {
	switch m.MType {
	case "gauge":
		s.SetGauge(m.ID, *m.Value)
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting gauge", zap.Error(err))
			return m, err
		}
		*m.Value = val
	case "counter":
		s.SetCounter(m.ID, *m.Delta)
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		*m.Delta = val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}

func (s *MemStorage) GetMetrics(m metrics.Metrics) (metrics.Metrics, error) {
	switch m.MType {
	case "gauge":
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while getting gauge", zap.Error(err))
			return m, err
		}
		*m.Value = val
	case "counter":
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		*m.Delta = val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}
