package handlers

import (
	"context"
	"errors"
	servicePackage "go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go-market/pkg/lunh"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type OrderLoadingHandler struct {
	service OrderLoadingService
	logger  *logging.ZapLogger
}

type OrderLoadingService interface {
	RegisterOrder(ctx context.Context, userID int, orderNumber string) error
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
	orderNumber := string(body)
	if !lunh.Validate(orderNumber) {
		h.logger.DebugCtx(r.Context(), "Invalid order number", zap.String("body", orderNumber))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	userID, err := userIDFromCtx(r.Context())
	if err != nil {
		h.logger.ErrorCtx(r.Context(), failedToRecoverUserIDErrorMessage, zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.service.RegisterOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, servicePackage.ErrOrderRegistered):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, servicePackage.ErrOrderRegisteredByAnotherUser):
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
