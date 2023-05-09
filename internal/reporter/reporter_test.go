package reporter

import (
	"github.com/sgladkov/harvester/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockConnection struct {
}

func (c *MockConnection) UpdateMetrics(_ *models.Metrics) error {
	return nil
}

func TestMetrics(t *testing.T) {
	c := MockConnection{}
	m := NewReporter(&c)
	require.Equal(t, 1, len(m.counters))
	require.Contains(t, m.counters, "PollCount")
	require.Equal(t, int64(0), m.counters["PollCount"])
	require.Equal(t, 0, len(m.gauges))
	require.NoError(t, m.Poll())
	require.Equal(t, 1, len(m.counters))
	require.Contains(t, m.counters, "PollCount")
	require.Equal(t, int64(1), m.counters["PollCount"])
	require.Equal(t, 28, len(m.gauges))
	require.NoError(t, m.Report())
}
