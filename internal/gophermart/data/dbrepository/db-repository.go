package dbrepository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	invalidUserID = -1
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

func (db *DBRepository) InsertUser(ctx context.Context, login, password string) (userID int, err error) {
	err = db.storage.QueryValue(ctx, insertUserQuery, []any{login, password}, []any{&userID})
	if err != nil {
		return invalidUserID, handleSQLError(err)
	}
	return userID, nil
}

//go:embed sql/validate_user.sql
var validateUserQuery string

func (db *DBRepository) ValidateUser(ctx context.Context, login, password string) (userID int, err error) {
	result := struct {
		userID          int
		passwordMatches bool
	}{}
	err = db.storage.QueryValue(
		ctx,
		validateUserQuery,
		[]any{login, password},
		[]any{&result.userID, &result.passwordMatches},
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return invalidUserID, data.ErrInvalidLogin
		default:
			return invalidUserID, fmt.Errorf("failed to validate user: %w", err)
		}
	}
	if !result.passwordMatches {
		return invalidUserID, data.ErrInvalidPassword
	}
	return result.userID, nil
}

//go:embed sql/insert_order.sql
var insertOrderQuery string

func (db *DBRepository) InsertOrder(ctx context.Context, order *data.Order) error {
	_, err := db.storage.Exec(
		ctx,
		insertOrderQuery,
		order.OrderNumber,
		string(order.Status),
		order.UserID,
		order.Accrual,
		order.UploadTime,
	)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_order_owner.sql
var selectOrderOwnerQuery string

func (db *DBRepository) GetOrderOwner(ctx context.Context, orderNumber string) (userID int, err error) {
	err = db.storage.QueryValue(ctx, selectOrderOwnerQuery, []any{orderNumber}, []any{&userID})
	if err != nil {
		return invalidUserID, handleSQLError(err)
	}
	return userID, nil
}

//go:embed sql/select_orders.sql
var selectOrdersQuery string

func (db *DBRepository) GetAllUserOrders(ctx context.Context, userID int) ([]data.Order, error) {
	rows, err := db.storage.Query(ctx, selectOrdersQuery, userID)
	if err != nil {
		return nil, handleSQLError(err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return make([]data.Order, 0), nil
		default:
			return nil, handleSQLError(err)
		}
	}

	result := make([]data.Order, 0)
	for rows.Next() {
		order := data.Order{
			UserID: userID,
		}
		err := rows.Scan(
			&order.OrderNumber,
			&order.Accrual,
			&order.UploadTime,
			&order.Status,
		)
		if err != nil {
			return nil, handleSQLError(err)
		}
		result = append(result, order)
	}
	return result, nil
}

func (db *DBRepository) GetOrders(
	ctx context.Context,
	limit int,
	allowedStatuses ...data.Status,
) ([]data.Order, error) {
	query := "SELECT number, user_id, accrual, upload_time, status FROM orders"
	if len(allowedStatuses) > 0 {
		query += fmt.Sprintf(" WHERE status IN (%s)", formatParams(2, len(allowedStatuses)))
	}
	if limit > 0 {
		query += " LIMIT $1"
	}
	args := make([]any, 0)
	if limit > 0 {
		args = append(args, limit)
	}
	for _, allowedStatus := range allowedStatuses {
		args = append(args, string(allowedStatus))
	}
	rows, err := db.storage.Query(ctx, query, args...)
	if err != nil {
		return nil, handleSQLError(err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return make([]data.Order, 0), nil
		default:
			return nil, handleSQLError(err)
		}
	}

	result := make([]data.Order, 0)
	for rows.Next() {
		var order data.Order
		err := rows.Scan(
			&order.OrderNumber,
			&order.UserID,
			&order.Accrual,
			&order.UploadTime,
			&order.Status,
		)
		if err != nil {
			return nil, handleSQLError(err)
		}
		result = append(result, order)
	}
	return result, nil
}

//go:embed sql/select_user_balance.sql
var selectUserBalanceQuery string

func (db *DBRepository) GetUserBalance(ctx context.Context, userID int) (decimal.Decimal, error) {
	var t *decimal.Decimal
	err := db.storage.QueryValue(ctx, selectUserBalanceQuery, []any{userID}, []any{&t})
	if err != nil {
		return decimal.Decimal{}, handleSQLError(err)
	}
	if t == nil {
		return decimal.Zero, nil
	}
	return *t, nil
}

//go:embed sql/update_user_balance.sql
var updateUserBalanceQuery string

func (db *DBRepository) SetUserBalance(ctx context.Context, userID int, value decimal.Decimal) error {
	_, err := db.storage.Exec(ctx, updateUserBalanceQuery, userID, value)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_order.sql
var selectOrderQuery string

func (db *DBRepository) GetOrder(ctx context.Context, orderNumber string) (userID int, status data.Status, err error) {
	db.logger.DebugCtx(ctx, "getting order", zap.String("orderNumber", orderNumber))
	err = db.storage.QueryValue(ctx, selectOrderQuery, []any{orderNumber}, []any{&userID, &status})
	if err != nil {
		return invalidUserID, data.NullStatus, handleSQLError(err)
	}
	return
}

//go:embed sql/update_order_status.sql
var updateOrderStatusQuery string

func (db *DBRepository) SetOrderStatus(
	ctx context.Context,
	orderNumber string,
	accrual decimal.Decimal,
	status data.Status,
) error {
	_, err := db.storage.Exec(ctx, updateOrderStatusQuery, orderNumber, status, accrual)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_user_withdrawals_sum.sql
var selectUserWithdrawalsSumQuery string

func (db *DBRepository) GetTotalUserWithdraw(ctx context.Context, userID int) (value decimal.Decimal, err error) {
	var t *decimal.Decimal
	err = db.storage.QueryValue(ctx, selectUserWithdrawalsSumQuery, []any{userID}, []any{&t})
	if err != nil {
		return decimal.Zero, handleSQLError(err)
	}
	if t == nil {
		return decimal.Zero, nil
	}
	return *t, nil
}

//go:embed sql/insert_withdrawal.sql
var insertWithdrawalQuery string

func (db *DBRepository) InsertWithdrawal(ctx context.Context, withdrawal data.Withdrawal) error {
	_, err := db.storage.Exec(
		ctx,
		insertWithdrawalQuery,
		withdrawal.OrderNumber,
		withdrawal.UserID,
		withdrawal.Amount,
		withdrawal.ProcessTime,
	)
	if err != nil {
		return handleSQLError(err)
	}
	return nil
}

//go:embed sql/select_withdrawals.sql
var selectWithdrawalsQuery string

func (db *DBRepository) GetAllUserWithdrawals(ctx context.Context, userID int) ([]data.Withdrawal, error) {
	rows, err := db.storage.Query(ctx, selectWithdrawalsQuery, userID)
	if err != nil {
		return nil, handleSQLError(err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return make([]data.Withdrawal, 0), nil
		default:
			return nil, handleSQLError(err)
		}
	}

	result := make([]data.Withdrawal, 0)
	for rows.Next() {
		order := data.Withdrawal{
			UserID: userID,
		}
		err := rows.Scan(
			&order.OrderNumber,
			&order.Amount,
			&order.ProcessTime,
		)
		if err != nil {
			return nil, handleSQLError(err)
		}
		result = append(result, order)
	}
	return result, nil
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
