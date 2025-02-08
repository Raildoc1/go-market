package dbrepository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"math/big"
	"strings"
)

const (
	invalidUserId = -1
	invalidPoints = -1
)

type DBStorage interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryValue(ctx context.Context, query string, args []any, dest []any) error
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
	err = db.storage.QueryValue(ctx, insertUserQuery, []any{login, password}, []any{&userId})
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
	err = db.storage.QueryValue(
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
	err = db.storage.QueryValue(ctx, selectOrderOwnerQuery, []any{orderNumber}, []any{&userId})
	if err != nil {
		return invalidUserId, handleSQLError(err)
	}
	return userId, nil
}

func (db *DBRepository) GetOrders(ctx context.Context, limit int, allowedStatuses ...data.Status) ([]string, error) {
	query := "SELECT number FROM orders"
	if len(allowedStatuses) > 0 {
		query += fmt.Sprintf(" WHERE status IN (%s)", formatParams(2, len(allowedStatuses)))
	}
	if limit > 0 {
		query += " LIMIT $1"
	}
	args := make([]any, len(allowedStatuses)+1)
	args[0] = limit
	for i, allowedStatus := range allowedStatuses {
		args[i+1] = string(allowedStatus)
	}
	rows, err := db.storage.Query(ctx, query, args...)
	if err != nil {
		return nil, handleSQLError(err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return make([]string, 0), nil
		default:
			return nil, handleSQLError(err)
		}
	}

	result := make([]string, 0)
	for rows.Next() {
		var orderNumber string
		err := rows.Scan(&orderNumber)
		if err != nil {
			return nil, handleSQLError(err)
		}
		result = append(result, orderNumber)
	}
	return result, nil
}

//go:embed sql/select_user_points.sql
var selectUserPointsQuery string

func (db *DBRepository) GetBonusPoints(ctx context.Context, userId int) (points int64, err error) {
	db.logger.DebugCtx(ctx, "getting points for user", zap.Int("user_id", userId))
	err = db.storage.QueryValue(ctx, selectUserPointsQuery, []any{userId}, []any{&points})
	if err != nil {
		return invalidPoints, handleSQLError(err)
	}
	return points, nil
}

//go:embed sql/update_user_points.sql
var updateUserPointsQuery string

func (db *DBRepository) SetBonusPoints(ctx context.Context, userId int, value int64) error {
	db.logger.DebugCtx(ctx, "getting points for user", zap.Int("user_id", userId), zap.Int64("value", value))
	_, err := db.storage.Exec(ctx, updateUserPointsQuery, userId, value)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_order.sql
var selectOrderQuery string

func (db *DBRepository) GetOrder(ctx context.Context, orderNumber string) (userId int, status data.Status, err error) {
	db.logger.DebugCtx(ctx, "getting order", zap.String("orderNumber", orderNumber))
	err = db.storage.QueryValue(ctx, selectOrderQuery, []any{orderNumber}, []any{&userId, &status})
	if err != nil {
		return invalidUserId, data.NullStatus, handleSQLError(err)
	}
	return
}

//go:embed sql/update_order_status.sql
var updateOrderStatusQuery string

func (db *DBRepository) SetOrderStatus(ctx context.Context, orderNumber string, status data.Status) error {
	_, err := db.storage.Exec(ctx, updateOrderStatusQuery, orderNumber, status)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
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

func formatParams(firstNumber, valuesCount int) string {
	currentNum := firstNumber
	values := make([]string, valuesCount)
	for i := range valuesCount {
		values[i] = fmt.Sprintf("$%v", currentNum)
		currentNum++
	}
	return strings.Join(values, ",")
}
