package service

import "context"

type Repository interface {
	InsertUser(ctx context.Context, login, password string) error
	ValidateUser(ctx context.Context, login, password string) error
}

type TransactionManager interface {
	DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type TokenFactory interface {
	Generate(login string) (string, error)
}
