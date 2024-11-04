package storage

import (
	"context"
	"database/sql"
)

type Driver interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
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
