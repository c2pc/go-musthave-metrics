package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Driver interface {
	Ping(ctx context.Context) error
	Close() error
}

type DB struct {
	*sqlx.DB
}

func Connect(ctx context.Context, dsn string) (Driver, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &DB{db}, err
}

func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db *DB) Close() error {
	return db.DB.Close()
}
