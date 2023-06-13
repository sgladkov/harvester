package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sgladkov/harvester/internal/config"
	"github.com/sgladkov/harvester/internal/httprouter"
	"github.com/sgladkov/harvester/internal/interfaces"
	"github.com/sgladkov/harvester/internal/logger"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"github.com/sgladkov/harvester/internal/utils"
	"go.uber.org/zap"
)

var storage interfaces.Storage
var db *sql.DB

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
	if len(*config.DatabaseDSN) > 0 {
		logger.Log.Info("Trying to open database", zap.String("DSN", *config.DatabaseDSN))
		err := utils.RetryOnError(
			func() error {
				db, err = sql.Open("postgres", *config.DatabaseDSN)
				if err != nil {
					return err
				}
				return nil
			},
			func(err error) bool {
				var pgErr *pgconn.PgError
				return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
			},
		)
		if err != nil {
			logger.Log.Fatal("Failed to open database", zap.Error(err))
		}
		defer func() {
			err := db.Close()
			if err != nil {
				logger.Log.Warn("failed to close database", zap.Error(err))
			}
		}()

		err = utils.RetryOnError(
			func() error {
				storage, err = storage2.NewPgStorage(db, true)
				return err
			},
			func(err error) bool {
				var pgErr *pgconn.PgError
				return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
			},
		)
		if err != nil {
			logger.Log.Fatal("Failed to create PgStorage", zap.Error(err))
		}
	} else {
		storage = storage2.NewMemStorage(*config.FileStorage, saveSettingsOnChange)
	}
	if *config.RestoreFlag {
		err := utils.RetryOnError(
			func() error {
				return storage.Read()
			},
			func(err error) bool {
				return errors.Is(err, os.ErrPermission)
			},
		)
		if err != nil {
			logger.Log.Warn("failed to read initial metrics values", zap.Error(err))
		}
	}

	if *config.StoreInterval > 0 {
		storeTicker := time.NewTicker(time.Duration(*config.StoreInterval) * time.Second)
		defer storeTicker.Stop()
		go func() {
			for range storeTicker.C {
				err := utils.RetryOnError(
					func() error {
						return storage.Save()
					},
					func(err error) bool {
						var pgErr *pgconn.PgError
						return errors.Is(err, os.ErrPermission) ||
							(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code))
					},
				)
				if err != nil {
					logger.Log.Warn("Failed to save metrics", zap.Error(err))
				} else {
					logger.Log.Info("Metrics are saved")
				}
			}
		}()
	}

	logger.Log.Info("Starting server", zap.String("address", *config.Endpoint))
	err = http.ListenAndServe(*config.Endpoint, httprouter.MetricsRouter(storage, db, config.Key))
	if err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
	err = utils.RetryOnError(
		func() error {
			return storage.Save()
		},
		func(err error) bool {
			var pgErr *pgconn.PgError
			return errors.Is(err, os.ErrPermission) ||
				(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code))
		},
	)
	if err != nil {
		logger.Log.Fatal("failed to store metrics", zap.Error(err))
	}
}
