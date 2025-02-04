package service

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/gophermart/data"
	"math/big"
)

var (
	ErrOrderRegisteredByAnotherUser = errors.New("order is already registered by another user")
	ErrOrderRegistered              = errors.New("order is already registered")
)

type Orders struct {
	transactionManager TransactionManager
	orderRepository    OrderRepository
}

type OrderRepository interface {
	InsertOrder(ctx context.Context, orderNumber *big.Int, userId int, status data.Status) error
	GetOrderOwner(ctx context.Context, orderNumber *big.Int) (userId int, err error)
}

func NewOrders(transactionManager TransactionManager, orderRepository OrderRepository) *Orders {
	return &Orders{
		transactionManager: transactionManager,
		orderRepository:    orderRepository,
	}
}

func (o *Orders) RegisterOrder(ctx context.Context, userId int, orderNumber *big.Int) error {
	err := o.orderRepository.InsertOrder(ctx, orderNumber, userId, data.NewStatus)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUniqueConstraintViolation):
			owner, err := o.orderRepository.GetOrderOwner(ctx, orderNumber)
			if err != nil {
				return fmt.Errorf("error checking order owner: %w", err)
			}
			if owner == userId {
				return ErrOrderRegistered
			}
			return ErrOrderRegisteredByAnotherUser
		default:
			return fmt.Errorf("error inserting order: %w", err)
		}
	}
	return nil
}
