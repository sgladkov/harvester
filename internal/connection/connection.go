package connection

import "github.com/sgladkov/harvester/internal/metrics"

type ServerConnection interface {
	UpdateMetrics(m *metrics.Metrics) error
}


