package main

import (
	"flag"
	"github.com/sgladkov/harvester/internal/httprouter"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var storage interfaces.Storage

func main() {
	// check arguments
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	logLevel := flag.String("l", "info", "log level (fatal,  error,  warn, info, debug)")
	storeInterval := flag.Int("i", 300, "metrics store interval")
	fileStorage := flag.String("s", "/tmp/metrics-db.json", "file to store ans restore metrics")
	restoreFlag := flag.Bool("r", true, "should server read initial metrics value from the file")
	flag.Parse()

	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*endpoint = address
	}
	envLogLevel := os.Getenv("LOG_LEVEL")
	if len(envLogLevel) > 0 {
		*logLevel = envLogLevel
	}

	err := logger.Initialize(*logLevel)
	if err != nil {
		log.Fatal(err)
	}

	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if len(envStoreInterval) > 0 {
		val, err := strconv.ParseInt(envStoreInterval, 10, 32)
		if err != nil {
			logger.Log.Fatal("Failed to interpret STORE_INTERVAL environment variable",
				zap.String("value", envStoreInterval), zap.Error(err))
		}
		*storeInterval = int(val)
	}
	envFileStorage := os.Getenv("FILE_STORAGE_PATH")
	if len(envFileStorage) > 0 {
		*fileStorage = envFileStorage
	}
	envRestore, exist := os.LookupEnv("RESTORE")
	if exist {
		val, err := strconv.ParseBool(envRestore)
		if err != nil {
			logger.Log.Fatal("Failed to interpret RESTORE environment variable",
				zap.String("value", envRestore), zap.Error(err))
		}
		*restoreFlag = val
	}

	saveSettingsOnChange := *storeInterval == 0
	storage = storage2.NewMemStorage(*fileStorage, saveSettingsOnChange)
	if *restoreFlag {
		err := storage.Read()
		if err != nil {
			logger.Log.Warn("failed to read initial metrics values from file", zap.Error(err))
		}
	}

	if *storeInterval > 0 {
		storeTicker := time.NewTicker(time.Duration(*storeInterval) * time.Second)
		defer storeTicker.Stop()
		go func() {
			for range storeTicker.C {
				err := storage.Save()
				if err != nil {
					logger.Log.Warn("Failed to save metrics", zap.Error(err))
				}
				logger.Log.Info("Metrics are read")
			}
		}()
	}

	logger.Log.Info("Starting server", zap.String("address", *endpoint))
	err = http.ListenAndServe(*endpoint, httprouter.MetricsRouter(storage))
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
	err = storage.Save()
	if err != nil {
		logger.Log.Fatal("failed to store metrics", zap.Error(err))
	}
}
