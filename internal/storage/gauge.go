package storage

import (
	"errors"
	"fmt"
)

type GaugeMetricValue float64

func (v GaugeMetricValue) String() string {
	return fmt.Sprintf("%f", v)
}

type GaugeStorage struct {
	storage map[string]GaugeMetricValue
}

func NewGaugeStorage() *GaugeStorage {
	return &GaugeStorage{
		storage: make(map[string]GaugeMetricValue),
	}
}

func (s *GaugeStorage) GetName() string {
	return "gauge"
}

func (s *GaugeStorage) Get(key string) (string, error) {
	value, ok := s.storage[key]
	if !ok {
		return "", errors.New("not found")
	}

	return value.String(), nil
}

func (s *GaugeStorage) Set(key string, value GaugeMetricValue) error {
	s.storage[key] = value
	return nil
}
