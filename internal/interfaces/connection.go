package interfaces

type ServerConnection interface {
	UpdateMetrics(m *Metrics) error
}
