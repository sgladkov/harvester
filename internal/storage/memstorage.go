package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/models"
	"go.uber.org/zap"
	"os"
	"sync"
)

type MemStorage struct {
	Gauges       map[string]float64
	Counters     map[string]int64
	lock         sync.Mutex
	fileStorage  string
	saveOnChange bool
}

func NewMemStorage(fileStorage string, saveOnChange bool) *MemStorage {
	return &MemStorage{
		Gauges:       make(map[string]float64),
		Counters:     make(map[string]int64),
		fileStorage:  fileStorage,
		saveOnChange: saveOnChange,
	}
}

func (s *MemStorage) GetGauge(name string) (float64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.Gauges[name]
	if !exists {
		return 0.0, fmt.Errorf("no gauge [%s]", name)
	}
	return value, nil
}

func (s *MemStorage) SetGauge(name string, value float64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Gauges[name] = value
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}

func (s *MemStorage) GetCounter(name string) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.Counters[name]
	if !exists {
		return 0, fmt.Errorf("no counter [%s]", name)
	}
	return value, nil
}

func (s *MemStorage) SetCounter(name string, value int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Counters[name] += value
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}

func (s *MemStorage) GetAll() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	var res string
	for n, v := range s.Gauges {
		res += fmt.Sprintf("%s=%g\n", n, v)
	}
	for n, v := range s.Counters {
		res += fmt.Sprintf("%s=%d\n", n, v)
	}
	return res
}

func (s *MemStorage) SetMetrics(m models.Metrics) (models.Metrics, error) {
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

func (s *MemStorage) GetMetrics(m models.Metrics) (models.Metrics, error) {
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

func (s *MemStorage) Save() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.doSave()
}

func (s *MemStorage) doSave() error {
	f, err := os.Create(s.fileStorage)
	if err != nil {
		logger.Log.Error("failed to create file to save metrics", zap.String("file", s.fileStorage), zap.Error(err))
		return err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logger.Log.Error("failed to close file with saved metrics", zap.Error(err))
		}
	}()

	data, err := json.Marshal(s)
	if err != nil {
		logger.Log.Error("failed to save metrics", zap.Error(err))
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		logger.Log.Error("failed to save metrics", zap.Error(err))
		return err
	}

	logger.Log.Info("metrics are saved", zap.String("file", s.fileStorage))
	return nil
}

func (s *MemStorage) Read() error {
	f, err := os.Open(s.fileStorage)
	if err != nil {
		logger.Log.Error("failed to open file to read metrics", zap.String("file", s.fileStorage), zap.Error(err))
		return err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logger.Log.Error("failed to close file with read metrics", zap.Error(err))
		}
	}()

	info, err := f.Stat()
	if err != nil {
		logger.Log.Error("failed to get info about the file to read metrics",
			zap.String("file", s.fileStorage), zap.Error(err))
		return err
	}

	data := make([]byte, info.Size())

	_, err = f.Read(data)
	if err != nil {
		logger.Log.Error("failed to read metrics", zap.Error(err))
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	err = json.Unmarshal(data, s)
	if err != nil {
		logger.Log.Error("failed to decode data for metrics", zap.Error(err))
		return err
	}
	return nil
}

func (s *MemStorage) SetMetricsBatch(metricsBatch []models.Metrics) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, m := range metricsBatch {
		err := models.ValidateMetricsID(m.ID)
		if err != nil {
			logger.Log.Warn("invalid metrics ID", zap.Error(err))
			return err
		}

		switch m.MType {
		case "gauge":
			if m.Value == nil {
				logger.Log.Warn("invalid gauge value")
				return errors.New("invalid gauge value")
			}
			s.Gauges[m.ID] = *m.Value
		case "counter":
			if m.Delta == nil {
				logger.Log.Warn("invalid counter value")
				return errors.New("invalid counter value")
			}
			s.Counters[m.ID] = *m.Delta
		default:
			logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
			return errors.New("unknown metrics type")
		}
	}
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}
