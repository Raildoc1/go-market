package accrualsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-market/internal/common/accrualsystemprotocol"
)

var (
	ErrNoOrderFound = errors.New("no order found")
)

type Config struct {
	ServerAddress string
}
type AccrualSystem struct {
	cfg Config
}

func NewAccrualSystem(cfg Config) *AccrualSystem {
	return &AccrualSystem{
		cfg: cfg,
	}
}

func (as *AccrualSystem) GetOrderStatus(ctx context.Context, orderNumber string) (accrualsystemprotocol.Order, error) {
	url := fmt.Sprintf("http://%s/api/orders/{number}", as.cfg.ServerAddress)
	resp, err := resty.
		New().
		R().
		SetContext(ctx).
		SetQueryParam("number", orderNumber).
		Get(url)
	if err != nil {
		return accrualsystemprotocol.Order{}, err
	}
	statusCode := resp.StatusCode()
	switch statusCode {
	case 204:
		return accrualsystemprotocol.Order{}, ErrNoOrderFound
	case 200:
		res := accrualsystemprotocol.Order{}
		err := json.Unmarshal(resp.Body(), &res)
		if err != nil {
			return accrualsystemprotocol.Order{}, err
		}
		return res, nil
	default:
		return accrualsystemprotocol.Order{}, fmt.Errorf("unexpected status code %v", statusCode)
	}
}
