package service

import "context"

type BalanceInfo struct {
	Balance     int64
	Withdrawals int64
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userId int) (points int64, err error)
	GetAllUserWithdrawals(ctx context.Context, userId int) (value int64, err error)
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
		withdrawals, err := w.repository.GetAllUserWithdrawals(ctx, userId)
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
