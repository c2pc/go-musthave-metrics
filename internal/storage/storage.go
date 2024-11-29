package storage

import (
	"context"
	"database/sql"
)

type Driver interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Type string

const (
	TypeMemory Type = "memory"
	TypeDB     Type = "db"
)

func (t Type) String() string {
	return string(t)
}

func (t Type) IsValid() bool {
	switch t {
	case TypeMemory, TypeDB:
		return true
	default:
		return false
	}
}

type Valuer[T any] interface {
	GetKey() string
	GetValue() T
}

type Value[T any] struct {
	Key   string
	Value T
}

func (v Value[T]) GetKey() string {
	return v.Key
}

func (v Value[T]) GetValue() T {
	return v.Value
}
