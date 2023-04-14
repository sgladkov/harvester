package internal

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	lock     sync.Mutex
}

func NewMemStorage() MemStorage {
	result := MemStorage{}
	result.gauges = make(map[string]float64)
	result.counters = make(map[string]int64)
	return result
}

func (m MemStorage) GetGauge(name string) (float64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, exists := m.gauges[name]
	if !exists {
		return 0.0, fmt.Errorf("no gauge [%s]", name)
	}
	return value, nil
}

func (m MemStorage) SetGauge(name string, value float64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.gauges[name] = value
	return nil
}

func (m MemStorage) GetCounter(name string) (int64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, exists := m.counters[name]
	if !exists {
		return 0, fmt.Errorf("no counter [%s]", name)
	}
	return value, nil
}

func (m MemStorage) SetCounter(name string, value int64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.counters[name] += value
	return nil
}

func (m MemStorage) GetAll() string {
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
