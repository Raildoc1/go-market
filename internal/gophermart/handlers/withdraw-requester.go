package handlers

import (
	"context"
	"errors"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go-market/pkg/lunh"
	"net/http"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WithdrawRequesterHandler struct {
	service WithdrawRequesterService
	logger  *logging.ZapLogger
}

type WithdrawRequesterService interface {
	Withdraw(ctx context.Context, userID int, orderNumber string, amount decimal.Decimal) error
}

type WithdrawalRequest struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
}

func NewWithdrawRequesterHandler(
	service WithdrawRequesterService,
	logger *logging.ZapLogger,
) *WithdrawRequesterHandler {
	return &WithdrawRequesterHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WithdrawRequesterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), failedToRecoverUserIDErrorMessage, zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request, err := decodeJSON[WithdrawalRequest](r.Body)
	if err != nil {
		h.logger.DebugCtx(r.Context(), "input decoding error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !lunh.Validate(request.OrderNumber) {
		h.logger.DebugCtx(r.Context(), "Invalid order number", zap.String("body", request.OrderNumber))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.service.Withdraw(r.Context(), userID, request.OrderNumber, request.Amount)
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
