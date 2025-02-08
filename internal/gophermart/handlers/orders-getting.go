package handlers

import (
	"context"
	"encoding/json"
	"go-market/internal/common/clientprotocol"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
)

type OrderGettingHandler struct {
	service OrderGettingService
	logger  *logging.ZapLogger
}

type OrderGettingService interface {
	GetAllOrders(ctx context.Context) ([]clientprotocol.Order, error)
}

func NewGettingHandler(service OrderGettingService, logger *logging.ZapLogger) *OrderGettingHandler {
	return &OrderGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetAllOrders(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error getting orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res, err := json.Marshal(orders)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error marshalling orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error writing response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
