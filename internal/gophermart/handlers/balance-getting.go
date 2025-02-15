package handlers

import (
	"context"
	"encoding/json"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

type BalanceInfo struct {
	Balance     float64 `json:"current"`
	Withdrawals float64 `json:"withdrawn"`
}

type BalanceGettingHandler struct {
	service BalanceGettingService
	logger  *logging.ZapLogger
}

type BalanceGettingService interface {
	GetUserBalanceInfo(ctx context.Context, userID int) (service.BalanceInfo, error)
}

func NewBalanceGettingHandler(service BalanceGettingService, logger *logging.ZapLogger) *BalanceGettingHandler {
	return &BalanceGettingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *BalanceGettingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), failedToRecoverUserIDErrorMessage, zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	balanceInfo, err := h.service.GetUserBalanceInfo(r.Context(), userID)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to get user balance info", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	balance, _ := balanceInfo.Balance.Float64()
	withdrawals, _ := balanceInfo.Withdrawals.Float64()
	convertedBalanceInfo := BalanceInfo{
		Balance:     balance,
		Withdrawals: withdrawals,
	}
	res, err := json.Marshal(convertedBalanceInfo)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error marshalling balance info", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Error writing response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
