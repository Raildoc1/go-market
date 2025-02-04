package handlers

import (
	"context"
	"errors"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"io"
	"math/big"
	"net/http"
)

type OrderLoadingHandler struct {
	service OrderLoadingService
	logger  *logging.ZapLogger
}

type OrderLoadingService interface {
	RegisterOrder(ctx context.Context, userId int, orderNumber *big.Int) error
}

func NewOrderLoadingHandler(service OrderLoadingService, logger *logging.ZapLogger) *OrderLoadingHandler {
	return &OrderLoadingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderLoadingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r.Context(), r.Body, h.logger)
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	const bodyLimit = 1024
	if r.ContentLength > bodyLimit {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to read request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bodyStr := string(body)
	orderNumber, ok := new(big.Int).SetString(bodyStr, 10)
	if !ok {
		h.logger.ErrorCtx(r.Context(), "Failed to parse order number", zap.String("body", bodyStr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId, err := userIdFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "Failed to recover user id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.service.RegisterOrder(r.Context(), userId, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderRegistered):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, service.ErrOrderRegisteredByAnotherUser):
			h.logger.DebugCtx(r.Context(), "Failed to register order", zap.Error(err))
			w.WriteHeader(http.StatusConflict)
			return
		default:
			h.logger.ErrorCtx(r.Context(), "Failed to register order", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}
