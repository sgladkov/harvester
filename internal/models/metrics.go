package models

import (
	"errors"
	"strings"
	"unicode"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func notLetterOrDigit(r rune) bool {
	if unicode.IsLetter(r) {
		return false
	}
	if unicode.IsDigit(r) {
		return false
	}
	return true
}

func IsValidMetricsID(id string) error {
	if len(id) == 0 {
		return errors.New("empty metrics id")
	}
	if strings.IndexFunc(id, notLetterOrDigit) != -1 {
		return errors.New("metrics id should contain letters or digits only")
	}
	return nil
}
