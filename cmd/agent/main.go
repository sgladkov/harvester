package main

import (
	"flag"
	"github.com/sgladkov/harvester/internal/connection"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/reporter"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	err := logger.Initialize("info")
	if err != nil {
		log.Fatal(err)
	}
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
			logger.Log.Fatal("Failed to interpret REPORT_INTERVAL environment variable",
				zap.String("value", reportStr), zap.Error(err))
		}
		*reportInterval = int(val)
	}
	pollStr := os.Getenv("POLL_INTERVAL")
	if len(pollStr) > 0 {
		val, err := strconv.ParseInt(pollStr, 10, 32)
		if err != nil {
			logger.Log.Fatal("Failed to interpret POLL_INTERVAL environment variable",
				zap.String("value", pollStr), zap.Error(err))
		}
		*pollInterval = int(val)
	}

	// add default url scheme if required
	if !strings.HasPrefix(*endpoint, "http://") && !strings.HasPrefix(*endpoint, "https://") {
		*endpoint = "http://" + *endpoint
	}

	r := connection.NewRestyClient(*endpoint)
	m := reporter.NewReporter(r)
	pollTicker := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	defer pollTicker.Stop()
	go func() {
		for range pollTicker.C {
			err := m.Poll()
			if err != nil {
				logger.Log.Warn("Failed to poll", zap.Error(err))
			}
			logger.Log.Info("Metrics are read")
		}
	}()
	reportTicker := time.NewTicker(time.Duration(*reportInterval) * time.Second)
	defer reportTicker.Stop()
	go func() {
		for range reportTicker.C {
			err := m.Report()
			if err != nil {
				logger.Log.Warn("Failed to report", zap.Error(err))
			}
			logger.Log.Info("Metrics are reported")
		}
	}()
	//r := bufio.NewReader(os.Stdin)
	//fmt.Println("Press Enter to exit")
	//r.ReadLine()
	for {
		time.Sleep(time.Second)
	}

}
