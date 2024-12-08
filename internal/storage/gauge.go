package storage

import (
	"context"
	"errors"
	"strconv"
	"sync"
)

type GaugeStorage struct {
	storageType Type
	mu          sync.RWMutex
	storage     map[string]float64
	db          Driver
}

func NewGaugeStorage(storageType Type, db Driver) (*GaugeStorage, error) {
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
	} else {
		return 0, ErrNotFound
	}

	return value, nil
}

func (s *GaugeStorage) getFromMemory(key string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.storage[key]
	if !ok {
		return 0, ErrNotFound
	}

	return value, nil
}

func (s *GaugeStorage) GetString(ctx context.Context, key string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}

	str, err := s.toString(value)
	if err != nil {
		return "", errors.Join(err, ErrInvalidValue)
	}

	return str, nil
}

func (s *GaugeStorage) Set(ctx context.Context, values ...Valuer[float64]) error {
	if len(values) == 0 {
		return nil
	}

	switch s.storageType {
	case TypeDB:
		return s.saveInDB(ctx, values...)
	default:
		return s.saveInMemory(values...)
	}
}

func (s *GaugeStorage) saveInDB(ctx context.Context, values ...Valuer[float64]) error {
	db, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, value := range values {
		_, err = s.db.ExecContext(ctx,
			`INSERT INTO gauges (key,value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = excluded.value`, value.GetKey(), value.GetValue())
		if err != nil {
			_ = db.Rollback()
			return err
		}
	}

	return db.Commit()
}

func (s *GaugeStorage) saveInMemory(values ...Valuer[float64]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, value := range values {
		s.storage[value.GetKey()] = value.GetValue()
	}

	return nil
}

func (s *GaugeStorage) SetString(ctx context.Context, values ...Valuer[string]) error {
	vs := make([]Valuer[float64], len(values))
	for i, value := range values {
		val, err := s.parseString(value.GetValue())
		if err != nil {
			return errors.Join(err, ErrInvalidValue)
		}
		vs[i] = Value[float64]{Key: value.GetKey(), Value: val}
	}

	return s.Set(ctx, vs...)
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
	s.mu.RLock()
	defer s.mu.RUnlock()

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
			return nil, ErrInvalidValue
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
