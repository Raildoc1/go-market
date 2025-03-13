package pgxstorage

import (
	"context"
	"errors"
	"fmt"
	"go-market/pkg/timeutils"
	"time"

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
	pool               *pgxpool.Pool
	retryAttemptDelays []time.Duration
}

func New(dbFactory DBFactory, retryAttemptDelays []time.Duration) (*DBStorage, error) {
	db, err := dbFactory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	return &DBStorage{
		pool:               db,
		retryAttemptDelays: retryAttemptDelays,
	}, nil
}

func (s *DBStorage) Close() {
	s.pool.Close()
}

func (s *DBStorage) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return queryInternal[pgconn.CommandTag](
		ctx,
		s.retryAttemptDelays,
		func(ctx context.Context) (pgconn.CommandTag, error) {
			return s.pool.Exec(ctx, query, args...)
		},
		func(ctx context.Context, tx pgx.Tx) (pgconn.CommandTag, error) {
			return tx.Exec(ctx, query, args...)
		},
	)
}

func (s *DBStorage) QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error) {
	return queryInternal[pgx.Row](
		ctx,
		s.retryAttemptDelays,
		func(ctx context.Context) (pgx.Row, error) {
			return s.pool.QueryRow(ctx, query, args...), nil
		},
		func(ctx context.Context, tx pgx.Tx) (pgx.Row, error) {
			return tx.QueryRow(ctx, query, args...), nil
		},
	)
}

func (s *DBStorage) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return queryInternal[pgx.Rows](
		ctx,
		s.retryAttemptDelays,
		func(ctx context.Context) (pgx.Rows, error) {
			return s.pool.Query(ctx, query, args...)
		},
		func(ctx context.Context, tx pgx.Tx) (pgx.Rows, error) {
			return tx.Query(ctx, query, args...)
		},
	)
}

func (s *DBStorage) QueryValue(ctx context.Context, query string, args []any, dest []any) error {
	row, err := s.QueryRow(ctx, query, args...)
	if err != nil {
		return err
	}
	return row.Scan(dest...) //nolint:wrapcheck // unnecessary
}

func queryInternal[T any](
	ctx context.Context,
	retryAttemptDelays []time.Duration,
	woTx func(context.Context) (T, error),
	withTx func(context.Context, pgx.Tx) (T, error),
) (T, error) {
	tx, err := getTransaction(ctx)
	if err != nil {
		switch {
		case errors.Is(err, errNoTransaction):
			return queryWithRetry[T](
				ctx,
				retryAttemptDelays,
				woTx,
			)
		default:
			var def T
			return def, err
		}
	}
	return queryWithRetry[T](
		ctx,
		retryAttemptDelays,
		func(ctx context.Context) (T, error) {
			return withTx(ctx, tx)
		},
	)
}

func queryWithRetry[T any](
	ctx context.Context,
	retryAttemptDelays []time.Duration,
	query func(context.Context) (T, error),
) (T, error) {
	return timeutils.Retry[T](
		ctx,
		retryAttemptDelays,
		query,
		func(_ T, err error) bool {
			return needRetry(err)
		},
	)
}

func needRetry(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// connection error
		return pgErr.Code[1] == '8'
	}
	return false
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
