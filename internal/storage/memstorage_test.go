package storage

import (
	"github.com/sgladkov/harvester/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemStorage_SimpleGetSet(t *testing.T) {
	s := NewMemStorage("", false)
	require.NoError(t, s.SetGauge("testg", 1.1))
	require.NoError(t, s.SetCounter("testc", 1))
	require.NoError(t, s.SetCounter("testc", 1))
	gauge, err := s.GetGauge("testg")
	require.NoError(t, err)
	require.Equal(t, 1.1, gauge)
	counter, err := s.GetCounter("testc")
	require.NoError(t, err)
	require.Equal(t, int64(2), counter)
	_, err = s.GetCounter("testg")
	require.Error(t, err)
	_, err = s.GetGauge("testc")
	require.Error(t, err)
}

func TestMemStorage_MetricsGetSet(t *testing.T) {
	s := NewMemStorage("", false)
	value := 1.1
	delta := int64(1)
	m, err := s.SetMetrics(models.Metrics{ID: "testg", MType: "gauge", Value: &value})
	require.NoError(t, err)
	require.Equal(t, 1.1, *m.Value)
	m, err = s.SetMetrics(models.Metrics{ID: "testg", MType: "gauge", Value: &value})
	require.NoError(t, err)
	require.Equal(t, 1.1, *m.Value)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "counter", Delta: &delta})
	require.NoError(t, err)
	require.Equal(t, int64(1), *m.Delta)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "counter", Delta: &delta})
	require.NoError(t, err)
	require.Equal(t, int64(2), *m.Delta)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "counter", Value: &value})
	require.Error(t, err)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "gauge", Delta: &delta})
	require.Error(t, err)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "counter"})
	require.Error(t, err)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "gauge"})
	require.Error(t, err)
	m, err = s.SetMetrics(models.Metrics{ID: "testc", MType: "unknown", Delta: &delta, Value: &value})
	require.Error(t, err)

	gauge, err := s.GetGauge("testg")
	require.NoError(t, err)
	require.Equal(t, 1.1, gauge)
	counter, err := s.GetCounter("testc")
	require.NoError(t, err)
	require.Equal(t, int64(2), counter)
	_, err = s.GetCounter("testg")
	require.Error(t, err)
	_, err = s.GetGauge("testc")
	require.Error(t, err)

	m, err = s.GetMetrics(models.Metrics{ID: "testg", MType: "gauge"})
	require.NoError(t, err)
	require.Equal(t, 1.1, *m.Value)
	m, err = s.GetMetrics(models.Metrics{ID: "testc", MType: "counter"})
	require.NoError(t, err)
	require.Equal(t, int64(2), *m.Delta)

	m, err = s.GetMetrics(models.Metrics{ID: "testg", MType: "counter"})
	require.Error(t, err)
	m, err = s.GetMetrics(models.Metrics{ID: "testc", MType: "gauge"})
	require.Error(t, err)
	m, err = s.GetMetrics(models.Metrics{ID: "testg"})
	require.Error(t, err)
	m, err = s.GetMetrics(models.Metrics{ID: "testc"})
	require.Error(t, err)
}

func TestMemStorage_StoreRestore(t *testing.T) {
	s := NewMemStorage("./test.storage", false)
	require.NoError(t, s.SetGauge("testg", 1.1))
	require.NoError(t, s.SetCounter("testc", 1))
	require.NoError(t, s.SetCounter("testc", 1))
	s.Save()

	s = NewMemStorage("./test.storage", false)
	_, err := s.GetCounter("testc")
	require.Error(t, err)
	_, err = s.GetGauge("testg")
	require.Error(t, err)

	s.Read()
	gauge, err := s.GetGauge("testg")
	require.NoError(t, err)
	require.Equal(t, 1.1, gauge)
	counter, err := s.GetCounter("testc")
	require.NoError(t, err)
	require.Equal(t, int64(2), counter)
	_, err = s.GetCounter("testg")
	require.Error(t, err)
	_, err = s.GetGauge("testc")
	require.Error(t, err)
}

func TestMemStorage_SaveOnChange(t *testing.T) {
	s := NewMemStorage("./test.storage", true)
	require.NoError(t, s.SetGauge("testg", 1.1))
	require.NoError(t, s.SetCounter("testc", 1))
	require.NoError(t, s.SetCounter("testc", 1))

	s = NewMemStorage("./test.storage", true)
	_, err := s.GetCounter("testc")
	require.Error(t, err)
	_, err = s.GetGauge("testg")
	require.Error(t, err)

	s.Read()
	gauge, err := s.GetGauge("testg")
	require.NoError(t, err)
	require.Equal(t, 1.1, gauge)
	counter, err := s.GetCounter("testc")
	require.NoError(t, err)
	require.Equal(t, int64(2), counter)
	_, err = s.GetCounter("testg")
	require.Error(t, err)
	_, err = s.GetGauge("testc")
	require.Error(t, err)
}
