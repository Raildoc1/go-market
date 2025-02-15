package handlers

import (
	"context"
	"go-market/internal/common/clientprotocol"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type OrderGettingHandler struct {
	service OrderGettingService
	logger  *logging.ZapLogger
}

type Order struct {
	UploadedAt time.Time                  `json:"uploaded_at"`
	Number     string                     `json:"number"`
	Status     clientprotocol.OrderStatus `json:"status"`
	Accrual    float64                    `json:"accrual"`
}

type OrderGettingService interface {
	GetAllOrders(ctx context.Context, userID int) ([]service.Order, error)
}

func NewOrderGettingHandler(service OrderGettingService, logger *logging.ZapLogger) *OrderGettingHandler {
	return &OrderGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), failedToRecoverUserIDErrorMessage, zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	orders, err := h.service.GetAllOrders(r.Context(), userID)
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
