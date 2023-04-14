package main

import (
	"flag"
	"fmt"
	"github.com/sgladkov/harvester/internal/metrics"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	pollInterval := flag.Int("p", 2, "poll interval")
	reportInterval := flag.Int("r", 10, "report interval")
	flag.Parse()
	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*endpoint = address
	}
	reportStr := os.Getenv("REPORT_INTERVAL")
	if len(reportStr) > 0 {
		val, err := strconv.ParseInt(reportStr, 10, 32)
		if err != nil {
			fmt.Printf("Failed to interpret REPORT_INTERVAL[%s] environment variable: %s\n", reportStr, err)
			return
		}
		*reportInterval = int(val)
	}
	pollStr := os.Getenv("POLL_INTERVAL")
	if len(pollStr) > 0 {
		val, err := strconv.ParseInt(pollStr, 10, 32)
		if err != nil {
			fmt.Printf("Failed to interpret POLL_INTERVAL[%s] environment variable: %s\n", pollStr, err)
			return
		}
		*pollInterval = int(val)
	}

	// add default url scheme if required
	if !strings.HasPrefix(*endpoint, "http://") && !strings.HasPrefix(*endpoint, "https://") {
		*endpoint = "http://" + *endpoint
	}

	m := metrics.NewMetrics(*endpoint)
	pollTicker := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	defer pollTicker.Stop()
	go func() {
		for range pollTicker.C {
			err := m.Poll()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics are read")
		}
	}()
	reportTicker := time.NewTicker(time.Duration(*reportInterval) * time.Second)
	defer reportTicker.Stop()
	go func() {
		for range reportTicker.C {
			err := m.Report()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics are reported")
		}
	}()
	//r := bufio.NewReader(os.Stdin)
	//fmt.Println("Press Enter to exit")
	//r.ReadLine()
	for {
	}

}
