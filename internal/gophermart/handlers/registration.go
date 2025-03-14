package handlers

import (
	"context"
	"errors"
	servicePackage "go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

type RegisterHandler struct {
	service RegistrationService
	logger  *logging.ZapLogger
}

type RegistrationInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegistrationService interface {
	Register(ctx context.Context, login string, password string) (string, error)
}

func NewRegisterHandler(service RegistrationService, logger *logging.ZapLogger) *RegisterHandler {
	return &RegisterHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r.Context(), r.Body, h.logger)

	input, err := decodeJSON[RegistrationInput](r.Body)
	if err != nil {
		h.logger.DebugCtx(r.Context(), "input decoding error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tkn, err := h.service.Register(r.Context(), input.Login, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, servicePackage.ErrLoginTaken):
			h.logger.DebugCtx(r.Context(), err.Error(), zap.String("login", input.Login))
			w.WriteHeader(http.StatusConflict)
			return
		default:
			h.logger.ErrorCtx(r.Context(), "registration service error", zap.Error(err), zap.Any("input", input))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Authorization", "Bearer "+tkn)
}
