package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMetrics(t *testing.T) {
	m := NewMetricsReporter("server_url")
	require.Equal(t, m.server, "server_url")
	require.Equal(t, 1, len(m.counters))
	require.Contains(t, m.counters, "PollCount")
	require.Equal(t, int64(0), m.counters["PollCount"])
	require.Equal(t, 0, len(m.gauges))
	require.NoError(t, m.Poll())
	require.Equal(t, 1, len(m.counters))
	require.Contains(t, m.counters, "PollCount")
	require.Equal(t, int64(1), m.counters["PollCount"])
	require.Equal(t, 28, len(m.gauges))
	require.Error(t, m.Report())
}
