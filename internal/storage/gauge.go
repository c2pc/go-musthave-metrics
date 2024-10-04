package storage

import (
	"errors"
	"strconv"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/handler"
)

type GaugeStorage struct {
	mu      sync.Mutex
	storage map[string]float64
}

func NewGaugeStorage() handler.Storager[float64] {
	return &GaugeStorage{
		storage: make(map[string]float64),
	}
}

func (s *GaugeStorage) GetName() string {
	return "gauge"
}

func (s *GaugeStorage) Get(key string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *GaugeStorage) GetString(key string) (string, error) {
	value, err := s.Get(key)
	if err != nil {
		return "", err
	}

	return s.toString(value)
}

func (s *GaugeStorage) Set(key string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value

	return nil
}

func (s *GaugeStorage) SetString(key string, value string) error {
	val, err := s.parseString(value)
	if err != nil {
		return err
	}

	return s.Set(key, val)
}

func (s *GaugeStorage) GetAll() (map[string]float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.storage, nil
}

func (s *GaugeStorage) GetAllString() (map[string]string, error) {
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

func (s *GaugeStorage) toString(value float64) (string, error) {
	return strconv.FormatFloat(value, 'f', -1, 64), nil
}

func (s *GaugeStorage) parseString(value string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
