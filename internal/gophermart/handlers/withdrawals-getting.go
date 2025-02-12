package handlers

import (
	"context"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
)

type WithdrawalsGettingHandler struct {
	service WithdrawalsGettingService
	logger  *logging.ZapLogger
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
	if err := tryWriteResponseJSON(w, withdrawals); err != nil {
		h.logger.ErrorCtx(r.Context(), "Error writing response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
