package dbrepository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
)

type DBStorage interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

type DBRepository struct {
	storage DBStorage
	logger  *logging.ZapLogger
}

func New(storage DBStorage, logger *logging.ZapLogger) *DBRepository {
	return &DBRepository{
		storage: storage,
		logger:  logger,
	}
}

func (db *DBRepository) InsertUser(ctx context.Context, login, password string) error {
	const query = `
		INSERT INTO users (login, password)
		VALUES ($1, crypt($2, gen_salt('md5')))`
	_, err := db.storage.Exec(ctx, query, login, password)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBRepository) ValidateUser(ctx context.Context, login, password string) error {
	const query = `
		SELECT (password = crypt($2, password)) 
		AS password_match
		FROM users
		WHERE login=$1;`
	row, err := db.storage.QueryRow(ctx, query, login, password)
	if err != nil {
		return err
	}
	passwordMatches := false
	err = row.Scan(&passwordMatches)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return data.ErrInvalidLogin
		default:
			return err
		}
	}
	if !passwordMatches {
		return data.ErrInvalidPassword
	}
	return nil
}
