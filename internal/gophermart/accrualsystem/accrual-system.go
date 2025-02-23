package accrualsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-market/internal/common/accrualsystemprotocol"
	"go-market/pkg/logging"
	"go-market/pkg/threadsafe"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	ErrNoOrderFound    = errors.New("no order found")
	ErrTooManyRequests = errors.New("too many requests")
)

type Config struct {
	ServerAddress string
}

type AccrualSystem struct {
	logger                 *logging.ZapLogger
	cfg                    Config
	remoteServiceAwakeTime *threadsafe.Time
}

func NewAccrualSystem(cfg Config, logger *logging.ZapLogger) *AccrualSystem {
	return &AccrualSystem{
		cfg:                    cfg,
		logger:                 logger,
		remoteServiceAwakeTime: threadsafe.NewTime(time.Now()),
	}
}

func (as *AccrualSystem) GetServiceAwakeTime() time.Time {
	return as.remoteServiceAwakeTime.Get()
}

func (as *AccrualSystem) GetOrderStatus(ctx context.Context, orderNumber string) (accrualsystemprotocol.Order, error) {
	if time.Now().Before(as.remoteServiceAwakeTime.Get()) {
		return accrualsystemprotocol.Order{}, ErrTooManyRequests
	}
	url := as.cfg.ServerAddress + "/api/orders/{number}"
	resp, err := resty.
		New().
		R().
		SetContext(ctx).
		SetPathParam("number", orderNumber).
		Get(url)
	if err != nil {
		return accrualsystemprotocol.Order{}, fmt.Errorf("get request failed: %w", err)
	}
	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusNoContent:
		as.logger.DebugCtx(ctx, "No order found")
		return accrualsystemprotocol.Order{}, ErrNoOrderFound
	case http.StatusOK:
		as.logger.DebugCtx(ctx, "Order found")
		res := accrualsystemprotocol.Order{}
		err := json.Unmarshal(resp.Body(), &res)
		if err != nil {
			as.logger.ErrorCtx(ctx, "Error unmarshalling order response", zap.Error(err))
			return accrualsystemprotocol.Order{}, fmt.Errorf("error unmarshalling order response: %w", err)
		}
		as.logger.DebugCtx(ctx, "Order found", zap.Any("order", res))
		return res, nil
	case http.StatusTooManyRequests:
		as.logger.DebugCtx(ctx, "Too many requests")
		retryAfterSeconds, err := strconv.Atoi(resp.Header().Get("Retry-After"))
		if err != nil {
			return accrualsystemprotocol.Order{}, fmt.Errorf("error converting retry-after header: %w", err)
		}
		retryAfter := time.Duration(retryAfterSeconds) * time.Second
		newRemoveServiceAwakeTime := time.Now().Add(retryAfter)
		as.remoteServiceAwakeTime.SetIf(
			newRemoveServiceAwakeTime,
			func(current time.Time) bool {
				return newRemoveServiceAwakeTime.After(current)
			},
		)
		return accrualsystemprotocol.Order{}, ErrTooManyRequests
	default:
		return accrualsystemprotocol.Order{}, fmt.Errorf("unexpected status code %v", statusCode)
	}
}
