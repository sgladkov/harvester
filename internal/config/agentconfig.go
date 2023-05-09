package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AgentConfig struct {
	Endpoint       *string
	PollInterval   *int
	ReportInterval *int
}

func (ac *AgentConfig) Read() error {
	ac.Endpoint = flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	ac.PollInterval = flag.Int("p", 2, "poll interval")
	ac.ReportInterval = flag.Int("r", 10, "report interval")
	flag.Parse()

	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*ac.Endpoint = address
	}
	reportStr := os.Getenv("REPORT_INTERVAL")
	if len(reportStr) > 0 {
		val, err := strconv.ParseInt(reportStr, 10, 32)
		if err != nil {
			return fmt.Errorf("failed to interpret REPORT_INTERVAL (=%s) environment variable, error is [%s]",
				reportStr, err)
		}
		*ac.ReportInterval = int(val)
	}
	pollStr := os.Getenv("POLL_INTERVAL")
	if len(pollStr) > 0 {
		val, err := strconv.ParseInt(pollStr, 10, 32)
		if err != nil {
			return fmt.Errorf("failed to interpret POLL_INTERVAL (=%s) environment variable, error is [%s]",
				pollStr, err)
		}
		*ac.PollInterval = int(val)
	}

	// add default url scheme if required
	if !strings.HasPrefix(*ac.Endpoint, "http://") && !strings.HasPrefix(*ac.Endpoint, "https://") {
		*ac.Endpoint = "http://" + *ac.Endpoint
	}

	return nil
}
