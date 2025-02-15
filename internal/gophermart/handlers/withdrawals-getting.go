package handlers

import (
	"context"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type WithdrawalsGettingHandler struct {
	service WithdrawalsGettingService
	logger  *logging.ZapLogger
}

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Amount      float64   `json:"sum"`
	ProcessTime time.Time `json:"processed_at"`
}

type WithdrawalsGettingService interface {
	GetAllUserWithdrawals(ctx context.Context, userId int) ([]service.Withdrawal, error)
}

func NewWithdrawalsGettingHandler(service WithdrawalsGettingService, logger *logging.ZapLogger) *WithdrawalsGettingHandler {
	return &WithdrawalsGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WithdrawalsGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to recover user id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	withdrawals, err := h.service.GetAllUserWithdrawals(r.Context(), userId)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error getting orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res := make([]Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		amount, _ := withdrawal.Amount.Float64()
		res[i] = Withdrawal{
			OrderNumber: withdrawal.OrderNumber,
			Amount:      amount,
			ProcessTime: withdrawal.ProcessTime,
		}
	}
	if err := tryWriteResponseJSON(w, res); err != nil {
		h.logger.ErrorCtx(r.Context(), "Error writing response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
