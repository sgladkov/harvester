package internal

type Storage interface {
	GetGauge(name string) (float64, error)
	SetGauge(name string, value float64) error
	GetCounter(name string) (int64, error)
	SetCounter(name string, value int64) error
}
