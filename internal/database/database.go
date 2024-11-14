package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
)

type DB struct {
	DB *sql.DB
}

func New(dsn string) (*DB, error) {
	db, err := connect(dsn)
	if err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func connect(dsn string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *DB) Ping() error {

	return db.DB.Ping()
}

func (db *DB) Close() error {
	if db.DB == nil {
		return nil
	}

	return db.DB.Close()
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := db.DB.ExecContext(ctx, query, args...)

	var rows int64
	if err == nil {
		rows, _ = result.RowsAffected()
	}

	logger.Log.Info("DB Exec", logger.Any("query", query), logger.Any("args", args), logger.Any("rows", rows), logger.Error(err))

	return result, err
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	result, err := db.DB.QueryContext(ctx, query, args...)

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
