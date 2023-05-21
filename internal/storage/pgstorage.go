package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sgladkov/harvester/internal/logger"
	"github.com/sgladkov/harvester/internal/models"
	"go.uber.org/zap"
	"sync"
)

type PgStorage struct {
	Gauges       map[string]float64
	Counters     map[string]int64
	lock         sync.Mutex
	db           *sql.DB
	saveOnChange bool
}

func NewPgStorage(databaseDSN string, saveOnChange bool) (*PgStorage, error) {
	logger.Log.Info("Trying to open database", zap.String("DSN", databaseDSN))
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		logger.Log.Error("Failed to open database", zap.Error(err))
		return nil, err
	}
	err = initDB(db)
	if err != nil {
		logger.Log.Error("Failed to init database", zap.Error(err))
		return nil, err
	}
	return &PgStorage{
		Gauges:       make(map[string]float64),
		Counters:     make(map[string]int64),
		db:           db,
		saveOnChange: saveOnChange,
	}, nil
}

func (s *PgStorage) GetGauge(name string) (float64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.Gauges[name]
	if !exists {
		return 0.0, fmt.Errorf("no gauge [%s]", name)
	}
	return value, nil
}

func (s *PgStorage) SetGauge(name string, value float64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Gauges[name] = value
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}

func (s *PgStorage) GetCounter(name string) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.Counters[name]
	if !exists {
		return 0, fmt.Errorf("no counter [%s]", name)
	}
	return value, nil
}

func (s *PgStorage) SetCounter(name string, value int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Counters[name] += value
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}

func (s *PgStorage) GetAll() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	var res string
	for n, v := range s.Gauges {
		res += fmt.Sprintf("%s=%g\n", n, v)
	}
	for n, v := range s.Counters {
		res += fmt.Sprintf("%s=%d\n", n, v)
	}
	return res
}

func (s *PgStorage) SetMetrics(m models.Metrics) (models.Metrics, error) {
	switch m.MType {
	case "gauge":
		if m.Value == nil {
			logger.Log.Warn("invalid gauge value")
			return m, errors.New("invalid gauge value")
		}
		err := s.SetGauge(m.ID, *m.Value)
		if err != nil {
			logger.Log.Warn("error while setting gauge", zap.Error(err))
			return m, err
		}
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting gauge", zap.Error(err))
			return m, err
		}
		m.Value = &val
	case "counter":
		if m.Delta == nil {
			logger.Log.Warn("invalid counter value")
			return m, errors.New("invalid counter value")
		}
		err := s.SetCounter(m.ID, *m.Delta)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		m.Delta = &val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}

func (s *PgStorage) GetMetrics(m models.Metrics) (models.Metrics, error) {
	switch m.MType {
	case "gauge":
		val, err := s.GetGauge(m.ID)
		if err != nil {
			logger.Log.Warn("error while getting gauge", zap.Error(err))
			return m, err
		}
		m.Value = &val
	case "counter":
		val, err := s.GetCounter(m.ID)
		if err != nil {
			logger.Log.Warn("error while setting counter", zap.Error(err))
			return m, err
		}
		m.Delta = &val
	default:
		logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
		return m, errors.New("unknown metrics type")
	}
	return m, nil
}

func (s *PgStorage) Save() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.doSave()
}

func (s *PgStorage) doSave() error {
	tx, err := s.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create db transaction", zap.Error(err))
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			logger.Log.Info("Failed to rollback db transaction", zap.String("error", err.Error()))
		}
	}()

	stmtGauges, err := tx.Prepare("INSERT INTO Gauges (id, value) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET value=EXCLUDED.value")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtGauges.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	stmtCounters, err := tx.Prepare("INSERT INTO Counters (id, value) VALUES ($1, $2) " +
		"ON CONFLICT (id) DO UPDATE SET value=EXCLUDED.value")
	if err != nil {
		logger.Log.Error("Failed to prepare query", zap.Error(err))
		return err
	}
	defer func() {
		err = stmtCounters.Close()
		if err != nil {
			logger.Log.Error("Failed to close statement", zap.Error(err))
		}
	}()

	for id, value := range s.Gauges {
		_, err = stmtGauges.Exec(id, value)
		if err != nil {
			logger.Log.Error("Failed to execute query", zap.Error(err))
			return err
		}
	}
	for id, value := range s.Counters {
		_, err = stmtCounters.Exec(id, value)
		if err != nil {
			logger.Log.Error("Failed to execute query", zap.Error(err))
			return err
		}
	}

	return tx.Commit()
}

func (s *PgStorage) Read() error {
	rowsGauges, err := s.db.Query("SELECT * FROM Gauges")
	if err != nil {
		logger.Log.Error("Failed to query Gauges data", zap.Error(err))
		return err
	}
	defer func() {
		err = rowsGauges.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()
	rowsCounters, err := s.db.Query("SELECT * FROM Counters")
	if err != nil {
		logger.Log.Error("Failed to query Counters data", zap.Error(err))
		return err
	}
	defer func() {
		err = rowsCounters.Close()
		if err != nil {
			logger.Log.Error("Failed to close rowset", zap.Error(err))
		}
	}()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.Gauges = make(map[string]float64)

	var id string
	var valueG float64
	for rowsGauges.Next() {
		err = rowsGauges.Scan(&id, &valueG)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return err
		}
		s.Gauges[id] = valueG
	}

	s.Counters = make(map[string]int64)
	var valueC int64
	for rowsCounters.Next() {
		err = rowsCounters.Scan(&id, &valueC)
		if err != nil {
			logger.Log.Error("Failed to get data from rowset", zap.Error(err))
			return err
		}
		s.Counters[id] = valueC
	}

	return nil
}

func initDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create db transaction", zap.Error(err))
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			logger.Log.Info("Failed to rollback db transaction", zap.String("error", err.Error()))
		}
	}()

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Gauges (id varchar(1024) PRIMARY KEY, value double precision)")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS Counters (id varchar(1024) PRIMARY KEY, value INTEGER)")
	if err != nil {
		logger.Log.Error("Failed to create db table", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (s *PgStorage) SetMetricsBatch(metricsBatch []models.Metrics) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, m := range metricsBatch {
		err := models.ValidateMetricsID(m.ID)
		if err != nil {
			logger.Log.Warn("invalid metrics ID", zap.Error(err))
			return err
		}

		switch m.MType {
		case "gauge":
			if m.Value == nil {
				logger.Log.Warn("invalid gauge value")
				return errors.New("invalid gauge value")
			}
			s.Gauges[m.ID] = *m.Value
		case "counter":
			if m.Delta == nil {
				logger.Log.Warn("invalid counter value")
				return errors.New("invalid counter value")
			}
			s.Counters[m.ID] = *m.Delta
		default:
			logger.Log.Warn("unknown metric type", zap.String("metric", m.MType))
			return errors.New("unknown metrics type")
		}
	}
	if s.saveOnChange {
		return s.doSave()
	}
	return nil
}
