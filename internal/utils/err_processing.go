package utils

import (
	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
	"time"
)

func RetryOnError(f func() error, errCheck func(error) bool) error {
	var count = 0
	timeouts := [...]int{1, 3, 5}
	for {
		err := f()
		if err != nil {
			if errCheck(err) {
				logger.Log.Warn("error, retry", zap.Error(err))
				time.Sleep(time.Duration(timeouts[count]) * time.Second)
				count++
				if count > 2 {
					return err
				}
				continue
			}
		}
		return nil
	}
}
