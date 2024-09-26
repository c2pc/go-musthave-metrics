package storage

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

type CounterStorage struct {
	mu      sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *CounterStorage) GetString(key string) (string, error) {
	value, err := s.Get(key)
	if err != nil {
		return "", err
	}

	return s.toString(value)
}

func (s *CounterStorage) Set(key string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.storage[key]; ok {
		s.storage[key] = val + value
	} else {
		s.storage[key] = value
	}
	return nil
}

func (s *CounterStorage) SetString(key string, value string) error {
	val, err := s.parseString(value)
	if err != nil {
		return err
	}

	return s.Set(key, val)
}

func (s *CounterStorage) GetAll() (map[string]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.storage, nil
}

func (s *CounterStorage) GetAllString() (map[string]string, error) {
	var response = make(map[string]string, len(s.storage))

	all, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	for k, v := range all {
		str, err := s.toString(v)
		if err != nil {
			return nil, err
		}
		response[k] = str
	}

	return response, nil
}

func (s *CounterStorage) toString(value int64) (string, error) {
	return fmt.Sprintf("%d", value), nil
}

func (s *CounterStorage) parseString(value string) (int64, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
