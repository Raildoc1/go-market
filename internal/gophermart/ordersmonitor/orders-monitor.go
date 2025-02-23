package ordersmonitor

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/common/accrualsystemprotocol"
	"go-market/internal/gophermart/accrualsystem"
	"go-market/internal/gophermart/data"
	"go-market/pkg/logging"
	"go-market/pkg/threadsafe"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type TransactionManager interface {
	DoWithTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type OrdersRepository interface {
	GetOrders(ctx context.Context, limit int, allowedStatuses ...data.Status) ([]data.Order, error)
	GetOrder(ctx context.Context, orderNumber string) (userID int, status data.Status, err error)
	SetOrderStatus(ctx context.Context, orderNumber string, accrual decimal.Decimal, status data.Status) error
}

type BonusPointsRepository interface {
	GetUserBalance(ctx context.Context, userID int) (decimal.Decimal, error)
	SetUserBalance(ctx context.Context, userID int, value decimal.Decimal) error
}

type AccrualSystem interface {
	GetServiceAwakeTime() time.Time
	GetOrderStatus(ctx context.Context, orderNumber string) (accrualsystemprotocol.Order, error)
}

type Config struct {
	TickPeriod        time.Duration
	WorkersCount      int
	TasksBufferLength int
}

type OrdersMonitor struct {
	orderStatusRepository OrdersRepository
	bonusPointsRepository BonusPointsRepository
	transactionManager    TransactionManager
	accrualSystem         AccrualSystem
	processingOrders      *threadsafe.HashSet[string]
	logger                *logging.ZapLogger
	done                  chan struct{}
	config                Config
}

func NewOrdersMonitor(
	config Config,
	orderStatusRepository OrdersRepository,
	bonusPointsRepository BonusPointsRepository,
	transactionManager TransactionManager,
	accrualSystem AccrualSystem,
	logger *logging.ZapLogger,
) *OrdersMonitor {
	return &OrdersMonitor{
		orderStatusRepository: orderStatusRepository,
		bonusPointsRepository: bonusPointsRepository,
		transactionManager:    transactionManager,
		accrualSystem:         accrualSystem,
		config:                config,
		processingOrders:      threadsafe.NewHashSet[string](),
		logger:                logger,
		done:                  make(chan struct{}),
	}
}

func (om *OrdersMonitor) Run() {
	orderNumbersChan := make(chan string, om.config.TasksBufferLength)

	wg := &sync.WaitGroup{}

	for range om.config.WorkersCount {
		wg.Add(1)
		go func(orderNumbersChan <-chan string) {
			defer wg.Done()
			om.worker(orderNumbersChan)
		}(orderNumbersChan)
	}

	wg.Add(1)
	go func(orderNumbersChan chan<- string) {
		defer wg.Done()
		om.scheduler(orderNumbersChan)
	}(orderNumbersChan)

	wg.Wait()
}

func (om *OrdersMonitor) Stop() {
	close(om.done)
}

func (om *OrdersMonitor) scheduler(orderNumbersChan chan<- string) {
	defer close(orderNumbersChan)

	ticker := time.NewTicker(om.config.TickPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-om.done:
			return
		case <-ticker.C:
			if err := om.tick(orderNumbersChan); err != nil {
				om.logger.ErrorCtx(context.Background(), "error while scheduling orders", zap.Error(err))
			}
		}
	}
}

func (om *OrdersMonitor) tick(orderNumbersChan chan<- string) error {
	maxTasksToSchedule := om.config.TasksBufferLength - len(orderNumbersChan)
	if maxTasksToSchedule <= 0 {
		return nil
	}
	orderNumbers, err := om.orderStatusRepository.GetOrders(
		context.Background(),
		maxTasksToSchedule,
		data.NewStatus,
		data.ProcessingStatus,
	)
	if err != nil {
		return fmt.Errorf("getting orders failed: %w", err)
	}
	if len(orderNumbers) == 0 {
		return nil
	}
	for _, order := range orderNumbers {
		orderNumber := order.OrderNumber
		if om.processingOrders.Contains(orderNumber) {
			continue
		}
		om.logger.DebugCtx(context.Background(), "scheduling order", zap.String("orderNumber", orderNumber))
		om.processingOrders.Add(orderNumber)
		orderNumbersChan <- orderNumber
	}
	return nil
}

func (om *OrdersMonitor) worker(orderNumberChan <-chan string) {
	for orderNumber := range orderNumberChan {
		err := om.handleOrder(orderNumber)
		om.processingOrders.Remove(orderNumber)
		if err != nil {
			om.logger.ErrorCtx(context.TODO(), "failed to handle order", zap.Error(err))
		}
	}
}

func (om *OrdersMonitor) handleOrder(orderNumber string) error {
	//nolint:wrapcheck // wrapping unnecessary
	return om.transactionManager.DoWithTransaction(context.Background(), func(ctx context.Context) error {
		userID, status, err := om.orderStatusRepository.GetOrder(ctx, orderNumber)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		switch status { //nolint:exhaustive // only these statuses considered as finished
		case data.ProcessedStatus:
			return nil
		case data.InvalidStatus:
			return nil
		case data.NullStatus:
			return errors.New("invalid order status")
		}
		remoteOrder, err := om.getRemoteOrder(ctx, orderNumber)
		if err != nil {
			switch {
			case errors.Is(err, accrualsystem.ErrNoOrderFound):
				return om.orderStatusRepository.SetOrderStatus(ctx, orderNumber, decimal.Zero, data.InvalidStatus)
			default:
				return fmt.Errorf("failed to get remote order status: %w", err)
			}
		}
		switch remoteOrder.Status {
		case accrualsystemprotocol.Invalid:
			return om.orderStatusRepository.SetOrderStatus(ctx, orderNumber, decimal.Zero, data.InvalidStatus)
		case accrualsystemprotocol.Processing, accrualsystemprotocol.Registered:
			return om.orderStatusRepository.SetOrderStatus(ctx, orderNumber, decimal.Zero, data.ProcessingStatus)
		case accrualsystemprotocol.Processed:
			currentPoints, err := om.bonusPointsRepository.GetUserBalance(ctx, userID)
			if err != nil {
				return fmt.Errorf("failed to get current bonus points: %w", err)
			}
			newBalance := currentPoints.Add(remoteOrder.Accrual)
			om.logger.DebugCtx(
				ctx,
				"setting balance",
				zap.Any("currentPoints", newBalance),
				zap.Any("newBalance", newBalance),
			)
			err = om.bonusPointsRepository.SetUserBalance(ctx, userID, newBalance)
			if err != nil {
				return fmt.Errorf("failed to set current bonus points: %w", err)
			}
			err = om.orderStatusRepository.SetOrderStatus(
				ctx,
				orderNumber,
				remoteOrder.Accrual,
				data.ProcessedStatus,
			)
			if err != nil {
				return fmt.Errorf("failed to set order status: %w", err)
			}
			return nil
		}
		return nil
	})
}

//nolint:wrapcheck // wrapping unnecessary
func (om *OrdersMonitor) getRemoteOrder(ctx context.Context, orderNumber string) (accrualsystemprotocol.Order, error) {
	for {
		if ctx.Err() != nil {
			return accrualsystemprotocol.Order{}, ctx.Err()
		}
		remoteOrder, err := om.accrualSystem.GetOrderStatus(ctx, orderNumber)
		if err != nil {
			switch {
			case errors.Is(err, accrualsystem.ErrTooManyRequests):
				timeToWait := time.Until(om.accrualSystem.GetServiceAwakeTime())
				select {
				case <-ctx.Done():
					return accrualsystemprotocol.Order{}, ctx.Err()
				case <-time.After(timeToWait):
					continue
				}
			default:
				return accrualsystemprotocol.Order{}, err
			}
		}
		return remoteOrder, nil
	}
}
