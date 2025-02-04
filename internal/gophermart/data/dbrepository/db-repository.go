package dbrepository

import (
	"context"
	_ "embed"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
	"math/big"
)

const (
	invalidUserId = -1
)

type DBStorage interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryValues(ctx context.Context, query string, args []any, dest []any) error
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

//go:embed sql/insert_user.sql
var insertUserQuery string

func (db *DBRepository) InsertUser(ctx context.Context, login, password string) (userId int, err error) {
	err = db.storage.QueryValues(ctx, insertUserQuery, []any{login, password}, []any{&userId})
	if err != nil {
		return invalidUserId, handleSQLError(err)
	}
	return userId, nil
}

//go:embed sql/validate_user.sql
var validateUserQuery string

func (db *DBRepository) ValidateUser(ctx context.Context, login, password string) (userId int, err error) {
	result := struct {
		userId          int
		passwordMatches bool
	}{}
	err = db.storage.QueryValues(
		ctx,
		validateUserQuery,
		[]any{login, password},
		[]any{&result.userId, &result.passwordMatches},
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return invalidUserId, data.ErrInvalidLogin
		default:
			return invalidUserId, err
		}
	}
	if !result.passwordMatches {
		return invalidUserId, data.ErrInvalidPassword
	}
	return result.userId, nil
}

//go:embed sql/insert_order.sql
var insertOrderQuery string

func (db *DBRepository) InsertOrder(ctx context.Context, orderNumber *big.Int, userId int, status data.Status) error {
	_, err := db.storage.Exec(ctx, insertOrderQuery, orderNumber.String(), string(status), userId)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_order_owner.sql
var selectOrderOwnerQuery string

func (db *DBRepository) GetOrderOwner(ctx context.Context, orderNumber *big.Int) (userId int, err error) {
	err = db.storage.QueryValues(ctx, selectOrderOwnerQuery, []any{orderNumber}, []any{&userId})
	if err != nil {
		return invalidUserId, handleSQLError(err)
	}
	return userId, nil
}

func handleSQLError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return data.ErrUniqueConstraintViolation
		}
	}
	return err
}
