package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"go-market/internal/common/clientprotocol"
	"go-market/internal/gophermart/data"
	"time"
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
	InsertOrder(ctx context.Context, order data.Order) error
	GetOrderOwner(ctx context.Context, orderNumber string) (userId int, err error)
	GetAllUserOrders(ctx context.Context, userId int) ([]data.Order, error)
}

func NewOrders(transactionManager TransactionManager, orderRepository OrderRepository) *Orders {
	return &Orders{
		transactionManager: transactionManager,
		orderRepository:    orderRepository,
	}
}

func (o *Orders) RegisterOrder(ctx context.Context, userId int, orderNumber string) error {
	order := data.Order{
		UserId:      userId,
		OrderNumber: orderNumber,
		Status:      data.NewStatus,
		Accrual:     decimal.Zero,
		UploadTime:  time.Now(),
	}
	err := o.orderRepository.InsertOrder(ctx, order)
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

func (o *Orders) GetAllOrders(ctx context.Context, userId int) ([]clientprotocol.Order, error) {
	orders, err := o.orderRepository.GetAllUserOrders(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting all orders: %w", err)
	}
	res := make([]clientprotocol.Order, len(orders))
	for i, order := range orders {
		protocolStatus, err := convert(order.Status)
		if err != nil {
			return nil, fmt.Errorf("error converting order: %w", err)
		}
		res[i] = clientprotocol.Order{
			Number:     order.OrderNumber,
			Status:     protocolStatus,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadTime,
		}
	}
	return res, nil
}

func convert(status data.Status) (clientprotocol.OrderStatus, error) {
	switch status {
	case data.NewStatus:
		return clientprotocol.New, nil
	case data.InvalidStatus:
		return clientprotocol.Invalid, nil
	case data.ProcessingStatus:
		return clientprotocol.Processing, nil
	case data.ProcessedStatus:
		return clientprotocol.Processed, nil
	}
	return clientprotocol.Null, fmt.Errorf("unknown status %s", status)
}
