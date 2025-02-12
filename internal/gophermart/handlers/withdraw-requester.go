package handlers

import (
	"context"
	"errors"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
)

type WithdrawRequesterHandler struct {
	service WithdrawRequesterService
	logger  *logging.ZapLogger
}

type WithdrawRequesterService interface {
	Withdraw(ctx context.Context, userId int, orderNumber string, amount int64) error
}

type WithdrawalRequest struct {
	OrderNumber string `json:"order"`
	Amount      int64  `json:"sum"`
}

func NewWithdrawRequesterHandler(service WithdrawRequesterService, logger *logging.ZapLogger) *WithdrawRequesterHandler {
	return &WithdrawRequesterHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WithdrawRequesterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to recover user id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := decodeJSON[WithdrawalRequest](r.Body)
	if err != nil {
		h.logger.DebugCtx(r.Context(), "input decoding error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.Withdraw(r.Context(), userId, request.OrderNumber, request.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotEnoughBalance):
			h.logger.DebugCtx(r.Context(), "", zap.Error(err))
			w.WriteHeader(http.StatusPaymentRequired)
			return
		default:
			h.logger.ErrorCtx(r.Context(), "Error getting orders", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
