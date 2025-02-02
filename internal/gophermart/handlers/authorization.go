package handlers

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/gophermart/services"
	"go.uber.org/zap"
	"net/http"
)

type AuthorizationHandler struct {
	service AuthorizationService
	logger  *zap.Logger
}

type AuthorizationInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthorizationService interface {
	Login(ctx context.Context, login string, password string) (string, error)
}

func NewAuthorizationHandler(service AuthorizationService, logger *zap.Logger) *AuthorizationHandler {
	return &AuthorizationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthorizationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := newRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)

	input, err := decodeJSON[RegistrationInput](r.Body)
	if err != nil {
		requestLogger.Debug("error decoding input", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tkn, err := h.service.Login(r.Context(), input.Login, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			requestLogger.Debug("invalid credentials", zap.String("login", input.Login))
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			requestLogger.Error("authorization handler error", zap.Error(err), zap.Any("input", input))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tkn))
}
