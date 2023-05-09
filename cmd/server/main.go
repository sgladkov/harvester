package main

import (
	"github.com/sgladkov/harvester/internal/config"
	"github.com/sgladkov/harvester/internal/httprouter"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

var storage interfaces.Storage

func main() {
	config := config.ServerConfig{}
	logLevel, err := config.Read()
	errLog := logger.Initialize(logLevel)
	if errLog != nil {
		log.Fatal(errLog)
	}
	if err != nil {
		log.Fatal(err)
	}

	saveSettingsOnChange := *config.StoreInterval == 0
	storage = storage2.NewMemStorage(*config.FileStorage, saveSettingsOnChange)
	if *config.RestoreFlag {
		err := storage.Read()
		if err != nil {
			logger.Log.Warn("failed to read initial metrics values from file", zap.Error(err))
		}
	}

	if *config.StoreInterval > 0 {
		storeTicker := time.NewTicker(time.Duration(*config.StoreInterval) * time.Second)
		defer storeTicker.Stop()
		go func() {
			for range storeTicker.C {
				err := storage.Save()
				if err != nil {
					logger.Log.Warn("Failed to save metrics", zap.Error(err))
				} else {
					logger.Log.Info("Metrics are saved")
				}
			}
		}()
	}

	logger.Log.Info("Starting server", zap.String("address", *config.Endpoint))
	err = http.ListenAndServe(*config.Endpoint, httprouter.MetricsRouter(storage))
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
	err = storage.Save()
	if err != nil {
		logger.Log.Fatal("failed to store metrics", zap.Error(err))
	}
}
