package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type contextKey int

const (
	transactionKey contextKey = iota
)

const (
	setupDatabaseRequest = `
		create table if not exists users
		(
			id       serial primary key,
			login    varchar(32) not null,
		    password varchar(128) not null, 
		);`
)

var errNoTransaction = errors.New("no transaction")

type DBFactory interface {
	Create() (*sql.DB, error)
}

type DBStorage struct {
	db *sql.DB
}

func New(dbFactory DBFactory) (*DBStorage, error) {
	db, err := dbFactory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	_, err = db.Exec(setupDatabaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return &DBStorage{
		db: db,
	}, nil
}

func (s *DBStorage) Close() error {
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (s *DBStorage) WithTransaction(ctx context.Context) (context.Context, *sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	ctxWithTransaction := context.WithValue(ctx, transactionKey, tx)
	return ctxWithTransaction, tx, nil
}

func (s *DBStorage) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.db.ExecContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
		default:
			return nil, err
		}
	}
	return tx.ExecContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func (s *DBStorage) QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.db.QueryRowContext(ctx, query, args...), nil
		default:
			return nil, err
		}
	}
	return tx.QueryRowContext(ctx, query, args...), nil
}

func (s *DBStorage) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.db.QueryContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
		default:
			return nil, err
		}
	}
	return tx.QueryContext(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func getTransaction(ctx context.Context) (*sql.Tx, error) {
	txVal := ctx.Value(transactionKey)
	if txVal == nil {
		return nil, errNoTransaction
	}
	tx, ok := txVal.(*sql.Tx)
	if !ok {
		return nil, errors.New("invalid transaction type")
	}
	return tx, nil
}
