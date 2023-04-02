package internal

type Storage interface {
	SetGauge(name string, value float64) error
	SetCounter(name string, value int64) error
}
