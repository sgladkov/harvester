package interfaces

import "github.com/sgladkov/harvester/internal/models"

type ServerConnection interface {
	UpdateMetrics(m *models.Metrics) error
	BatchUpdateMetrics(metricsBatch []models.Metrics) error
}
