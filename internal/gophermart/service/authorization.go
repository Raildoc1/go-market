package service

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/gophermart/data"
	"strconv"
)

var (
	ErrLoginTaken         = errors.New("login is already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	UserIDClaimName = "user_id"
)

type UserRepository interface {
	InsertUser(ctx context.Context, login, password string) (userID int, err error)
	ValidateUser(ctx context.Context, login, password string) (userID int, err error)
}

type TokenFactory interface {
	Generate(extraClaims map[string]string) (string, error)
}

type Authorization struct {
	userRepository     UserRepository
	transactionManager TransactionManager
	tokenFactory       TokenFactory
}

func NewAuthorization(
	userRepository UserRepository,
	transactionManager TransactionManager,
	tokenFactory TokenFactory,
) *Authorization {
	return &Authorization{
		userRepository:     userRepository,
		transactionManager: transactionManager,
		tokenFactory:       tokenFactory,
	}
}

func (r *Authorization) Register(ctx context.Context, login string, password string) (string, error) {
	userID, err := r.userRepository.InsertUser(ctx, login, password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUniqueConstraintViolation):
			return "", ErrLoginTaken
		default:
			return "", fmt.Errorf("error inserting user: %w", err)
		}
	}

	payload := map[string]string{
		UserIDClaimName: strconv.Itoa(userID),
	}
	token, err := r.tokenFactory.Generate(payload)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return token, nil
}

func (r *Authorization) Login(ctx context.Context, login string, password string) (string, error) {
	userID, err := r.userRepository.ValidateUser(ctx, login, password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidPassword):
			return "", ErrInvalidCredentials
		case errors.Is(err, data.ErrInvalidLogin):
			return "", ErrInvalidCredentials
		default:
			return "", fmt.Errorf("error inserting user: %w", err)
		}
	}

	payload := map[string]string{
		UserIDClaimName: strconv.Itoa(userID),
	}
	token, err := r.tokenFactory.Generate(payload)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return token, nil
}
