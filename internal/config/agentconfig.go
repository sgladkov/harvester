package config

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type AgentConfig struct {
	Endpoint       *string
	PollInterval   *int
	ReportInterval *int
	Key            *string
	RateLimit      *uint
}

func (ac *AgentConfig) Read() error {
	ac.Endpoint = flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	ac.PollInterval = flag.Int("p", 2, "poll interval")
	ac.ReportInterval = flag.Int("r", 10, "report interval")
	ac.Key = flag.String("k", "", "key to verify data integrity")
	ac.RateLimit = flag.Uint("l", 0, "max simultaneous server requests")
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
	key := os.Getenv("KEY")
	if len(key) > 0 {
		*ac.Key = key
	}
	rateLimitStr := os.Getenv("RATE_LIMIT")
	if len(rateLimitStr) > 0 {
		val, err := strconv.ParseUint(rateLimitStr, 10, 32)
		if err != nil {
			return fmt.Errorf("failed to interpret RATE_LIMIT (=%s) environment variable, error is [%s]",
				pollStr, err)
		}
		*ac.RateLimit = uint(val)
	}

	// if rate limit is 0 set it to cpu number:
	if *ac.RateLimit == 0 {
		*ac.RateLimit = uint(runtime.GOMAXPROCS(0))
	}
	// add default url scheme if required
	if !strings.HasPrefix(*ac.Endpoint, "http://") && !strings.HasPrefix(*ac.Endpoint, "https://") {
		*ac.Endpoint = "http://" + *ac.Endpoint
	}

	return nil
}
