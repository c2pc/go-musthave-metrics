package storage

import (
	"errors"
	"fmt"
)

type CounterMetricValue int64

func (v CounterMetricValue) String() string {
	return fmt.Sprintf("%d", v)
}

type CounterStorage struct {
	storage map[string]CounterMetricValue
}

func NewCounterStorage() *CounterStorage {
	return &CounterStorage{
		storage: make(map[string]CounterMetricValue),
	}
}

func (s *CounterStorage) GetName() string {
	return "counter"
}

func (s *CounterStorage) Get(key string) (string, error) {
	value, ok := s.storage[key]
	if !ok {
		return "", errors.New("not found")
	}

	return fmt.Sprintf("%d", value), nil
}

func (s *CounterStorage) Set(key string, value CounterMetricValue) error {
	if val, ok := s.storage[key]; ok {
		s.storage[key] = val + value
	} else {
		s.storage[key] = value
	}
	return nil
}
