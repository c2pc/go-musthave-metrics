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
	dsn string
	*sqlx.DB
}

func New(dsn string) Driver {
	return &DB{dsn: dsn}
}

func (db *DB) checkConn(ctx context.Context) error {
	var err error

	if db.DB == nil {
		db.DB, err = sqlx.ConnectContext(ctx, "postgres", db.dsn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.checkConn(ctx); err != nil {
		return err
	}

	return db.DB.PingContext(ctx)
}

func (db *DB) Close() error {
	if db.DB == nil {
		return nil
	}

	return db.DB.Close()
}
