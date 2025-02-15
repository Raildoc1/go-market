package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"time"
)

var (
	ErrNotEnoughBalance = errors.New("not enough balance")
)

type BalanceInfo struct {
	Balance     decimal.Decimal
	Withdrawals decimal.Decimal
}

type Withdrawal struct {
	OrderNumber string
	Amount      decimal.Decimal
	ProcessTime time.Time
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID int) (balance decimal.Decimal, err error)
	SetUserBalance(ctx context.Context, userID int, value decimal.Decimal) error
	GetTotalUserWithdraw(ctx context.Context, userID int) (value decimal.Decimal, err error)
	InsertWithdrawal(ctx context.Context, withdrawal data.Withdrawal) error
	GetAllUserWithdrawals(ctx context.Context, userID int) ([]data.Withdrawal, error)
}

type Wallet struct {
	transactionManager TransactionManager
	repository         BalanceRepository
	logger             *logging.ZapLogger
}

func NewWallet(transactionManager TransactionManager, repository BalanceRepository, logger *logging.ZapLogger) *Wallet {
	return &Wallet{
		transactionManager: transactionManager,
		repository:         repository,
		logger:             logger,
	}
}

func (w *Wallet) GetUserBalanceInfo(ctx context.Context, userID int) (BalanceInfo, error) {
	res := BalanceInfo{}
	err := w.transactionManager.DoWithTransaction(ctx, func(ctx context.Context) error {
		balance, err := w.repository.GetUserBalance(ctx, userID)
		if err != nil {
			return fmt.Errorf("getting user balance failed: %w", err)
		}
		res.Balance = balance
		withdrawals, err := w.repository.GetTotalUserWithdraw(ctx, userID)
		if err != nil {
			return fmt.Errorf("getting total user withdrawals failed: %w", err)
		}
		res.Withdrawals = withdrawals
		return nil
	})
	if err != nil {
		return BalanceInfo{}, err //nolint:wrapcheck // unnecessary
	}
	return res, nil
}

func (w *Wallet) Withdraw(ctx context.Context, userID int, orderNumber string, amount decimal.Decimal) error {
	w.logger.DebugCtx(
		ctx,
		"withdraw",
		zap.Int("userID", userID),
		zap.String("orderNumber", orderNumber),
		zap.String("amount", amount.String()),
	)
	return w.transactionManager.DoWithTransaction(ctx, func(ctx context.Context) error {
		balance, err := w.repository.GetUserBalance(ctx, userID)
		if err != nil {
			return fmt.Errorf("getting user balance failed: %w", err)
		}
		if balance.LessThan(amount) {
			return ErrNotEnoughBalance
		}
		newBalance := balance.Sub(amount)
		err = w.repository.SetUserBalance(ctx, userID, newBalance)
		if err != nil {
			return fmt.Errorf("setting user balance failed: %w", err)
		}
		return w.repository.InsertWithdrawal(ctx, data.Withdrawal{
			OrderNumber: orderNumber,
			Amount:      amount,
			UserID:      userID,
			ProcessTime: time.Now(),
		})
	})
}

func (w *Wallet) GetAllUserWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error) {
	withdrawals, err := w.repository.GetAllUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user withdrawals failed: %w", err)
	}
	res := make([]Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		res[i] = Withdrawal{
			OrderNumber: withdrawal.OrderNumber,
			Amount:      withdrawal.Amount,
			ProcessTime: withdrawal.ProcessTime,
		}
	}
	return res, nil
}
