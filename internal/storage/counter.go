package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/retry"
)

type CounterStorage struct {
	storageType Type
	mu          sync.RWMutex
	storage     map[string]int64
	db          Driver
}

func NewCounterStorage(storageType Type, db Driver) (*CounterStorage, error) {
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
	var rows *sql.Rows
	err := retry.Retry(
		func() error {
			var err error
			rows, err = s.db.QueryContext(ctx, `SELECT value FROM counters WHERE key=$1 LIMIT 1`, key)
			if err != nil {
				return err
			}

			return rows.Err()
		},
		func(err error) bool {
			return errors.Is(err, driver.ErrBadConn)
		},
		[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var value int64
	if rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	} else {
		return 0, errors.New("not found")
	}

	return value, nil
}

func (s *CounterStorage) getFromMemory(key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

func (s *CounterStorage) Set(ctx context.Context, values ...Valuer[int64]) error {
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

func (s *CounterStorage) saveInDB(ctx context.Context, values ...Valuer[int64]) error {
	db, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, value := range values {
		err := retry.Retry(
			func() error {
				_, err := s.db.ExecContext(ctx,
					`INSERT INTO counters (key,value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = counters.value + excluded.value`, value.GetKey(), value.GetValue())
				return err
			},
			func(err error) bool {
				return errors.Is(err, driver.ErrBadConn)
			},
			[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		)
		if err != nil {
			_ = db.Rollback()
			return err
		}
	}

	return db.Commit()
}

func (s *CounterStorage) saveInMemory(values ...Valuer[int64]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, value := range values {
		if val, ok := s.storage[value.GetKey()]; ok {
			s.storage[value.GetKey()] = val + value.GetValue()
		} else {
			s.storage[value.GetKey()] = value.GetValue()
		}
	}

	return nil
}

func (s *CounterStorage) SetString(ctx context.Context, values ...Valuer[string]) error {
	vs := make([]Valuer[int64], len(values))
	for i, value := range values {
		val, err := s.parseString(value.GetValue())
		if err != nil {
			return err
		}
		vs[i] = Value[int64]{Key: value.GetKey(), Value: val}
	}

	return s.Set(ctx, vs...)
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
	var rows *sql.Rows
	err := retry.Retry(
		func() error {
			var err error
			rows, err = s.db.QueryContext(ctx, `SELECT key, value FROM counters`)
			if err != nil {
				return err
			}

			return rows.Err()
		},
		func(err error) bool {
			return errors.Is(err, driver.ErrBadConn)
		},
		[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
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
	s.mu.RLock()
	defer s.mu.RLock()

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
