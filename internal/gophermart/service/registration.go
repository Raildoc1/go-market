package service

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/gophermart/data"
)

type Registration struct {
	repository         Repository
	transactionManager TransactionManager
	tokenFactory       TokenFactory
}

func NewRegistration(
	repository Repository,
	transactionManager TransactionManager,
	tokenFactory TokenFactory,
) *Registration {
	return &Registration{
		repository:         repository,
		transactionManager: transactionManager,
		tokenFactory:       tokenFactory,
	}
}

func (r *Registration) Register(ctx context.Context, login string, password string) (string, error) {
	err := r.repository.InsertUser(ctx, login, password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserAlreadyExists):
			return "", ErrLoginTaken
		default:
			return "", fmt.Errorf("error inserting user: %w", err)
		}
	}

	token, err := r.tokenFactory.Generate(login)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return token, nil
}
