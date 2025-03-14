package handlers

import (
	"context"
	"errors"
	servicePackage "go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

type AuthorizationHandler struct {
	service AuthorizationService
	logger  *logging.ZapLogger
}

type AuthorizationInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthorizationService interface {
	Login(ctx context.Context, login string, password string) (string, error)
}

func NewAuthorizationHandler(service AuthorizationService, logger *logging.ZapLogger) *AuthorizationHandler {
	return &AuthorizationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthorizationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r.Context(), r.Body, h.logger)

	input, err := decodeJSON[RegistrationInput](r.Body)
	if err != nil {
		h.logger.DebugCtx(r.Context(), "input decoding error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tkn, err := h.service.Login(r.Context(), input.Login, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, servicePackage.ErrInvalidCredentials):
			h.logger.DebugCtx(r.Context(), err.Error(), zap.Any("input", input))
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			h.logger.ErrorCtx(r.Context(), "login service error", zap.Error(err), zap.Any("input", input))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Authorization", "Bearer "+tkn)
}
