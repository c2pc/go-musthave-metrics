package storage

import (
	"errors"
)

type GaugeStorage struct {
	storage map[string]float64
}

func NewGaugeStorage() *GaugeStorage {
	return &GaugeStorage{
		storage: make(map[string]float64),
	}
}

func (s *GaugeStorage) GetName() string {
	return "counter"
}

func (s *GaugeStorage) Get(key string) (float64, error) {
	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *GaugeStorage) Set(key string, value float64) error {
	s.storage[key] = value
	return nil
}
