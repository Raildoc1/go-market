package pgxstorage

import (
	"context"
	"fmt"
)

type TransactionsManager struct {
	storage *DBStorage
}

func NewTransactionsManager(storage *DBStorage) *TransactionsManager {
	return &TransactionsManager{
		storage: storage,
	}
}

func (tm *TransactionsManager) DoWithTransaction(
	ctx context.Context,
	f func(ctx context.Context) error,
) error {
	ctxWithTransaction, tx, err := tm.storage.withTransaction(ctx)
	if err != nil {
		return err
	}
	err = f(ctxWithTransaction)
	if err != nil {
		rollbackErr := tx.Rollback(context.Background())
		if rollbackErr != nil {
			return fmt.Errorf("transaction rollback failed: %w, rollback caused by %w", rollbackErr, err)
		}
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		rollbackErr := tx.Rollback(context.Background())
		if rollbackErr != nil {
			return fmt.Errorf("transaction rollback failed: %w, rollback caused by %w", rollbackErr, err)
		}
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}
