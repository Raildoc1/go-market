package pgxstorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey int

const (
	transactionKey contextKey = iota
)

var errNoTransaction = errors.New("no transaction")

type DBFactory interface {
	Create() (*pgxpool.Pool, error)
}

type DBStorage struct {
	pool *pgxpool.Pool
}

func New(dbFactory DBFactory) (*DBStorage, error) {
	db, err := dbFactory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	return &DBStorage{
		pool: db,
	}, nil
}

func (s *DBStorage) Close() {
	s.pool.Close()
}

func (s *DBStorage) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.pool.Exec(ctx, query, args...) //nolint:wrapcheck // unnecessary
		default:
			return pgconn.CommandTag{}, err
		}
	}
	return tx.Exec(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func (s *DBStorage) QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.pool.QueryRow(ctx, query, args...), nil
		default:
			return nil, err
		}
	}
	return tx.QueryRow(ctx, query, args...), nil
}

func (s *DBStorage) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return s.pool.Query(ctx, query, args...) //nolint:wrapcheck // unnecessary
		default:
			return nil, err
		}
	}
	return tx.Query(ctx, query, args...) //nolint:wrapcheck // unnecessary
}

func (s *DBStorage) withTransaction(ctx context.Context) (context.Context, pgx.Tx, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	ctxWithTransaction := context.WithValue(ctx, transactionKey, tx)
	return ctxWithTransaction, tx, nil
}

func getTransaction(ctx context.Context) (pgx.Tx, error) {
	txVal := ctx.Value(transactionKey)
	if txVal == nil {
		return nil, errNoTransaction
	}
	tx, ok := txVal.(pgx.Tx)
	if !ok {
		return nil, errors.New("invalid transaction type")
	}
	return tx, nil
}
