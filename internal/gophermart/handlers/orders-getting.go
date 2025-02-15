package handlers

import (
	"context"
	"go-market/internal/common/clientprotocol"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type OrderGettingHandler struct {
	service OrderGettingService
	logger  *logging.ZapLogger
}

type Order struct {
	Number     string                     `json:"number"`
	Status     clientprotocol.OrderStatus `json:"status"`
	Accrual    float64                    `json:"accrual"`
	UploadedAt time.Time                  `json:"uploaded_at"`
}

type OrderGettingService interface {
	GetAllOrders(ctx context.Context, userId int) ([]service.Order, error)
}

func NewOrderGettingHandler(service OrderGettingService, logger *logging.ZapLogger) *OrderGettingHandler {
	return &OrderGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to recover user id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	orders, err := h.service.GetAllOrders(r.Context(), userId)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error getting orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res := make([]Order, len(orders))
	for i, order := range orders {
		accrual, _ := order.Accrual.Float64()
		res[i] = Order{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    accrual,
			UploadedAt: order.UploadedAt,
		}
	}
	if err := tryWriteResponseJSON(w, res); err != nil {
		h.logger.ErrorCtx(r.Context(), "Error writing response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
