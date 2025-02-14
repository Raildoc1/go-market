package service

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"go-market/internal/gophermart/data"
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
	GetUserBalance(ctx context.Context, userId int) (balance decimal.Decimal, err error)
	SetUserBalance(ctx context.Context, userId int, value decimal.Decimal) error
	GetTotalUserWithdraw(ctx context.Context, userId int) (value decimal.Decimal, err error)
	InsertWithdrawal(ctx context.Context, withdrawal data.Withdrawal) error
	GetAllUserWithdrawals(ctx context.Context, userId int) ([]data.Withdrawal, error)
}

type Wallet struct {
	transactionManager TransactionManager
	repository         BalanceRepository
}

func NewWallet(transactionManager TransactionManager, repository BalanceRepository) *Wallet {
	return &Wallet{
		transactionManager: transactionManager,
		repository:         repository,
	}
}

func (w *Wallet) GetUserBalanceInfo(ctx context.Context, userId int) (BalanceInfo, error) {
	res := BalanceInfo{}
	err := w.transactionManager.DoWithTransaction(ctx, func(ctx context.Context) error {
		balance, err := w.repository.GetUserBalance(ctx, userId)
		if err != nil {
			return err
		}
		res.Balance = balance
		withdrawals, err := w.repository.GetTotalUserWithdraw(ctx, userId)
		if err != nil {
			return err
		}
		res.Withdrawals = withdrawals
		return nil
	})
	if err != nil {
		return BalanceInfo{}, err
	}
	return res, nil
}

func (w *Wallet) Withdraw(ctx context.Context, userId int, orderNumber string, amount decimal.Decimal) error {
	return w.transactionManager.DoWithTransaction(ctx, func(ctx context.Context) error {
		balance, err := w.repository.GetUserBalance(ctx, userId)
		if err != nil {
			return err
		}
		if balance.LessThan(amount) {
			return ErrNotEnoughBalance
		}
		newBalance := balance.Sub(amount)
		err = w.repository.SetUserBalance(ctx, userId, newBalance)
		if err != nil {
			return err
		}
		return w.repository.InsertWithdrawal(ctx, data.Withdrawal{
			OrderNumber: orderNumber,
			Amount:      amount,
			UserId:      userId,
			ProcessTime: time.Now(),
		})
	})
}

func (w *Wallet) GetAllUserWithdrawals(ctx context.Context, userId int) ([]Withdrawal, error) {
	withdrawals, err := w.repository.GetAllUserWithdrawals(ctx, userId)
	if err != nil {
		return nil, err
	}
	var res []Withdrawal
	for _, withdrawal := range withdrawals {
		res = append(res, Withdrawal{
			OrderNumber: withdrawal.OrderNumber,
			Amount:      withdrawal.Amount,
			ProcessTime: withdrawal.ProcessTime,
		})
	}
	return res, nil
}
