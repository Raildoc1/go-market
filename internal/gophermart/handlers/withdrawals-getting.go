package handlers

import (
	"context"
	servicePackage "go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type WithdrawalsGettingHandler struct {
	service WithdrawalsGettingService
	logger  *logging.ZapLogger
}

type Withdrawal struct {
	ProcessTime time.Time `json:"processed_at"`
	OrderNumber string    `json:"order"`
	Amount      float64   `json:"sum"`
}

type WithdrawalsGettingService interface {
	GetAllUserWithdrawals(ctx context.Context, userID int) ([]servicePackage.Withdrawal, error)
}

func NewWithdrawalsGettingHandler(
	service WithdrawalsGettingService,
	logger *logging.ZapLogger,
) *WithdrawalsGettingHandler {
	return &WithdrawalsGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WithdrawalsGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), failedToRecoverUserIDErrorMessage, zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	withdrawals, err := h.service.GetAllUserWithdrawals(r.Context(), userID)
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
