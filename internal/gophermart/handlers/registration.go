package handlers

import (
	"context"
	"errors"
	"fmt"
	"go-market/internal/gophermart/services"
	"go.uber.org/zap"
	"net/http"
)

type RegisterHandler struct {
	service RegistrationService
	logger  *zap.Logger
}

type RegistrationInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegistrationService interface {
	Register(ctx context.Context, login string, password string) (string, error)
}

func NewRegisterHandler(service RegistrationService, logger *zap.Logger) *RegisterHandler {
	return &RegisterHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := newRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)

	input, err := decodeJSON[RegistrationInput](r.Body)
	if err != nil {
		requestLogger.Debug("error decoding input", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tkn, err := h.service.Register(r.Context(), input.Login, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrLoginTaken):
			requestLogger.Debug("login is already taken", zap.String("login", input.Login))
			w.WriteHeader(http.StatusConflict)
			return
		default:
			requestLogger.Error("registration handler error", zap.Error(err), zap.Any("input", input))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tkn))
}
