package internal

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() MemStorage {
	result := MemStorage{}
	result.gauges = make(map[string]float64)
	result.counters = make(map[string]int64)
	return result
}

func (m MemStorage) SetGauge(name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m MemStorage) SetCounter(name string, value int64) error {
	m.counters[name] += value
	return nil
}
