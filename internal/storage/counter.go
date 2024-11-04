package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/handler"
)

type CounterStorage struct {
	storageType Type
	mu          sync.Mutex
	storage     map[string]int64
	db          Driver
}

func NewCounterStorage(storageType Type, db Driver) (handler.Storager[int64], error) {
	if !storageType.IsValid() {
		return nil, errors.New("invalid storage type")
	}

	return &CounterStorage{
		storageType: storageType,
		storage:     make(map[string]int64),
		db:          db,
	}, nil
}

func (s *CounterStorage) GetName() string {
	return "counter"
}

func (s *CounterStorage) Get(ctx context.Context, key string) (int64, error) {
	switch s.storageType {
	case TypeDB:
		return s.getFromDB(ctx, key)
	default:
		return s.getFromMemory(key)
	}
}

func (s *CounterStorage) getFromDB(ctx context.Context, key string) (int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT value FROM counters WHERE key=$1 LIMIT 1`, key)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return 0, rows.Err()
	}

	var value int64
	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
		break
	}

	return value, nil
}

func (s *CounterStorage) getFromMemory(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.storage[key]
	if !ok {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *CounterStorage) GetString(ctx context.Context, key string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}

	return s.toString(value)
}

func (s *CounterStorage) Set(ctx context.Context, key string, value int64) error {
	switch s.storageType {
	case TypeDB:
		return s.saveInDB(ctx, key, value)
	default:
		return s.saveInMemory(key, value)
	}
}

func (s *CounterStorage) saveInDB(ctx context.Context, key string, value int64) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO counters (key,value) VALUES ($1,$2) ON CONFLICT (key) DO UPDATE SET value = counters.value + excluded.value`, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *CounterStorage) saveInMemory(key string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.storage[key]; ok {
		s.storage[key] = val + value
	} else {
		s.storage[key] = value
	}

	return nil
}

func (s *CounterStorage) SetString(ctx context.Context, key string, value string) error {
	val, err := s.parseString(value)
	if err != nil {
		return err
	}

	return s.Set(ctx, key, val)
}

func (s *CounterStorage) GetAll(ctx context.Context) (map[string]int64, error) {
	switch s.storageType {
	case TypeDB:
		return s.getAllFromDB(ctx)
	default:
		return s.getAllFromMemory()
	}
}

func (s *CounterStorage) getAllFromDB(ctx context.Context) (map[string]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM counters`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	result := make(map[string]int64)
	if rows.Next() {
		var key string
		var value int64
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

func (s *CounterStorage) getAllFromMemory() (map[string]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.storage, nil
}

func (s *CounterStorage) GetAllString(ctx context.Context) (map[string]string, error) {
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
