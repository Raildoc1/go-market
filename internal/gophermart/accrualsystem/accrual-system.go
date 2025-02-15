package accrualsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-market/internal/common/accrualsystemprotocol"
	"go-market/pkg/logging"
	"go.uber.org/zap"
)

var (
	ErrNoOrderFound = errors.New("no order found")
)

type Config struct {
	ServerAddress string
}
type AccrualSystem struct {
	logger *logging.ZapLogger
	cfg    Config
}

func NewAccrualSystem(cfg Config, logger *logging.ZapLogger) *AccrualSystem {
	return &AccrualSystem{
		cfg:    cfg,
		logger: logger,
	}
}

func (as *AccrualSystem) GetOrderStatus(ctx context.Context, orderNumber string) (accrualsystemprotocol.Order, error) {
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
	case 204:
		as.logger.DebugCtx(ctx, "No order found")
		return accrualsystemprotocol.Order{}, ErrNoOrderFound
	case 200:
		as.logger.DebugCtx(ctx, "Order found")
		res := accrualsystemprotocol.Order{}
		err := json.Unmarshal(resp.Body(), &res)
		if err != nil {
			as.logger.ErrorCtx(ctx, "Error unmarshalling order response", zap.Error(err))
			return accrualsystemprotocol.Order{}, fmt.Errorf("error unmarshalling order response: %w", err)
		}
		as.logger.DebugCtx(ctx, "Order found", zap.Any("order", res))
		return res, nil
	default:
		return accrualsystemprotocol.Order{}, fmt.Errorf("unexpected status code %v", statusCode)
	}
}
