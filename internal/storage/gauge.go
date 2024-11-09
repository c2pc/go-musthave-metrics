package storage

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/handler"
)

type GaugeStorage struct {
	storageType Type
	mu          sync.Mutex
	storage     map[string]float64
	db          Driver
}

func NewGaugeStorage(storageType Type, db Driver) (handler.Storager[float64], error) {
	if !storageType.IsValid() {
		return nil, errors.New("invalid storage type")
	}

	return &GaugeStorage{
		storageType: storageType,
		storage:     make(map[string]float64),
		db:          db,
	}, nil
}

func (s *GaugeStorage) GetName() string {
	return "gauge"
}

func (s *GaugeStorage) Get(ctx context.Context, key string) (float64, error) {
	switch s.storageType {
	case TypeDB:
		return s.getFromDB(ctx, key)
	default:
		return s.getFromMemory(key)
	}
}

func (s *GaugeStorage) getFromDB(ctx context.Context, key string) (float64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT value FROM gauges WHERE key=$1 LIMIT 1`, key)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return 0, rows.Err()
	}

	var value float64
	if rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	}

	return value, nil
}

func (s *GaugeStorage) getFromMemory(key string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *GaugeStorage) GetString(ctx context.Context, key string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return s.toString(value)
}

func (s *GaugeStorage) Set(ctx context.Context, key string, value float64) error {
	switch s.storageType {
	case TypeDB:
		return s.saveInDB(ctx, key, value)
	default:
		return s.saveInMemory(key, value)
	}
}

func (s *GaugeStorage) saveInDB(ctx context.Context, key string, value float64) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO gauges (key,value) VALUES ($1,$2) ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *GaugeStorage) saveInMemory(key string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value

	return nil
}

func (s *GaugeStorage) SetString(ctx context.Context, key string, value string) error {
	val, err := s.parseString(value)
	if err != nil {
		return err
	}

	return s.Set(ctx, key, val)
}

func (s *GaugeStorage) GetAll(ctx context.Context) (map[string]float64, error) {
	switch s.storageType {
	case TypeDB:
		return s.getAllFromDB(ctx)
	default:
		return s.getAllFromMemory()
	}
}

func (s *GaugeStorage) getAllFromDB(ctx context.Context) (map[string]float64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM gauges`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	result := make(map[string]float64)
	for rows.Next() {
		var key string
		var value float64
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

func (s *GaugeStorage) getAllFromMemory() (map[string]float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.storage, nil
}

func (s *GaugeStorage) GetAllString(ctx context.Context) (map[string]string, error) {
	var response = make(map[string]string, len(s.storage))

	all, err := s.GetAll(ctx)
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
