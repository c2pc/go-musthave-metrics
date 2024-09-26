package storage

import (
	"errors"
)

type CounterStorage struct {
	storage map[string]int64
}

func NewCounterStorage() *CounterStorage {
	return &CounterStorage{
		storage: make(map[string]int64),
	}
}

func (s *CounterStorage) GetName() string {
	return "counter"
}

func (s *CounterStorage) Get(key string) (int64, error) {
	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *CounterStorage) Set(key string, value int64) error {
	if val, ok := s.storage[key]; ok {
		s.storage[key] = val + value
	} else {
		s.storage[key] = value
	}
	return nil
}
