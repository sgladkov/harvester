package main

import (
	"log"
	"time"

	config2 "github.com/sgladkov/harvester/internal/config"
	"github.com/sgladkov/harvester/internal/connection"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/reporter"
	"github.com/sgladkov/harvester/internal/utils"
	"go.uber.org/zap"
)

func main() {
	err := logger.Initialize("info")
	if err != nil {
		log.Fatal(err)
	}

	config := config2.AgentConfig{}
	err = config.Read()
	if err != nil {
		logger.Log.Fatal("failed to read config params", zap.Error(err))
	}

	r := connection.NewRestyClient(*config.Endpoint)
	m := reporter.NewReporter(r)
	pollTicker := time.NewTicker(time.Duration(*config.PollInterval) * time.Second)
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
	reportTicker := time.NewTicker(time.Duration(*config.ReportInterval) * time.Second)
	defer reportTicker.Stop()
	go func() {
		for range reportTicker.C {
			err := utils.RetryOnError(
				func() error {
					return m.BatchReport()
				},
				func(err error) bool {
					return true
				},
			)
			if err != nil {
				logger.Log.Warn("Failed to report", zap.Error(err))
			}
			logger.Log.Info("Metrics are reported")
		}
	}()
	for {
		time.Sleep(time.Second)
	}
}
