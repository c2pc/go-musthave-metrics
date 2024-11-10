package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
)

type DB struct {
	dsn string
	DB  *sql.DB
}

func New(dsn string) *DB {
	return &DB{dsn: dsn}
}

func (db *DB) checkConn(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var err error

	if db.DB == nil {
		db.DB, err = sql.Open("postgres", db.dsn)
		if err != nil {
			return err
		}

		err = db.DB.PingContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Ping() error {
	if err := db.checkConn(context.Background()); err != nil {
		return err
	}

	return db.DB.Ping()
}

func (db *DB) Close() error {
	if db.DB == nil {
		return nil
	}

	return db.DB.Close()
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := func() (sql.Result, error) {
		if err := db.checkConn(ctx); err != nil {
			return nil, err
		}

		return db.DB.ExecContext(ctx, query, args...)
	}()

	var rows int64
	if err == nil {
		rows, _ = result.RowsAffected()
	}

	logger.Log.Info("DB Exec", logger.Any("query", query), logger.Any("args", args), logger.Any("rows", rows), logger.Error(err))

	return result, err
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	result, err := func() (*sql.Rows, error) {
		if err := db.checkConn(ctx); err != nil {
			return nil, err
		}

		return db.DB.QueryContext(ctx, query, args...)
	}()

	var rows int
	if err == nil {
		columns, err := result.Columns()
		if err == nil {
			rows = len(columns)
		}
	}

	logger.Log.Info("DB Query", logger.Any("query", query), logger.Any("args", args), logger.Any("rows", rows), logger.Error(err))

	return result, err
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, opts)
}
