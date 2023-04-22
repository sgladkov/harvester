package storage

import (
	"errors"
	"fmt"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type Storage interface {
	GetGauge(name string) (float64, error)
	SetGauge(name string, value float64) error
	GetCounter(name string) (int64, error)
	SetCounter(name string, value int64) error
	GetAll() string
	SetMetrics(m interfaces.Metrics) (interfaces.Metrics, error)
	GetMetrics(m interfaces.Metrics) (interfaces.Metrics, error)
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

func (s *MemStorage) GetGauge(name string) (float64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.gauges[name]
	if !exists {
		return 0.0, fmt.Errorf("no gauge [%s]", name)
	}
	return value, nil
}

func (s *MemStorage) SetGauge(name string, value float64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) GetCounter(name string) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.counters[name]
	if !exists {
		return 0, fmt.Errorf("no counter [%s]", name)
	}
	return value, nil
}

func (s *MemStorage) SetCounter(name string, value int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.counters[name] += value
	return nil
}

func (s *MemStorage) GetAll() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	var res string
	for n, v := range s.gauges {
		res += fmt.Sprintf("%s=%g\n", n, v)
	}
	for n, v := range s.counters {
		res += fmt.Sprintf("%s=%d\n", n, v)
	}
	return res
}

func (s *MemStorage) SetMetrics(m interfaces.Metrics) (interfaces.Metrics, error) {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			logger.Log.Warn("invalid gauge value")
			return m, errors.New("invalid gauge value")
		}
		err := s.SetGauge(m.ID, *m.Value)
		if err != nil {
			logger.Log.Warn("error while setting gauge", zap.Error(err))
			return m, err
		}
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting gauge", zap.Error(err))
			return m, err
		}
		m.Value = &val
	case "counter":
		if m.Delta == nil {
			logger.Log.Warn("invalid counter value")
			return m, errors.New("invalid counter value")
		}
		err := s.SetCounter(m.ID, *m.Delta)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		m.Delta = &val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}

func (s *MemStorage) GetMetrics(m interfaces.Metrics) (interfaces.Metrics, error) {
	switch m.MType {
	case "gauge":
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while getting gauge", zap.Error(err))
			return m, err
		}
		m.Value = &val
	case "counter":
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		m.Delta = &val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}
