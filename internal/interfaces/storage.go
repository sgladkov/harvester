package interfaces

import "github.com/sgladkov/harvester/internal/models"

type Storage interface {
	GetGauge(name string) (float64, error)
	SetGauge(name string, value float64) error
	GetCounter(name string) (int64, error)
	SetCounter(name string, value int64) error
	GetAll() string
	SetMetrics(m models.Metrics) (models.Metrics, error)
	GetMetrics(m models.Metrics) (models.Metrics, error)
	Save() error
	Read() error
}
