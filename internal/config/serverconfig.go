package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ServerConfig struct {
	Endpoint      *string
	LogLevel      *string
	StoreInterval *int
	FileStorage   *string
	RestoreFlag   *bool
	DatabaseDSN   *string
	Key           *string
}

func (sc *ServerConfig) Read() (string, error) {
	logLevel := "Info"
	sc.Endpoint = flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	sc.LogLevel = flag.String("l", "info", "log level (fatal,  error,  warn, info, debug)")
	logLevel = *sc.LogLevel
	sc.StoreInterval = flag.Int("i", 300, "metrics store interval")
	sc.FileStorage = flag.String("s", "/tmp/metrics-db.json", "file to store ans restore metrics")
	sc.RestoreFlag = flag.Bool("r", true, "should server read initial metrics value from the file")
	sc.DatabaseDSN = flag.String("d", "", "database connection string for PostgreSQL")
	sc.Key = flag.String("k", "", "key to verify data integrity")
	flag.Parse()

	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		*sc.Endpoint = address
	}
	envLogLevel := os.Getenv("LOG_LEVEL")
	if len(envLogLevel) > 0 {
		*sc.LogLevel = envLogLevel
		logLevel = envLogLevel
	}
	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if len(envStoreInterval) > 0 {
		val, err := strconv.ParseInt(envStoreInterval, 10, 32)
		if err != nil {
			return logLevel, fmt.Errorf("failed to interpret STORE_INTERVAL (=%s) environment variable, error is [%s]",
				envStoreInterval, err)
		}
		*sc.StoreInterval = int(val)
	}
	envFileStorage := os.Getenv("FILE_STORAGE_PATH")
	if len(envFileStorage) > 0 {
		*sc.FileStorage = envFileStorage
	}
	envRestore, exist := os.LookupEnv("RESTORE")
	if exist {
		val, err := strconv.ParseBool(envRestore)
		if err != nil {
			return logLevel, fmt.Errorf("failed to interpret RESTORE (=%s) environment variable, error is [%s]",
				envRestore, err)
		}
		*sc.RestoreFlag = val
	}
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	if len(envDatabaseDSN) > 0 {
		*sc.DatabaseDSN = envDatabaseDSN
	}
	key := os.Getenv("KEY")
	if len(key) > 0 {
		*sc.Key = key
	}

	return logLevel, nil
}
