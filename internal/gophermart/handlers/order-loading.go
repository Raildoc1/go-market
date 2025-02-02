package handlers

import (
	"go-market/pkg/logging"
	"net/http"
)

type OrderLoadingHandler struct {
	logger *logging.ZapLogger
}

func NewOrderLoadingHandler(logger *logging.ZapLogger) *OrderLoadingHandler {
	return &OrderLoadingHandler{
		logger: logger,
	}
}

func (h *OrderLoadingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer closeBody(r.Context(), r.Body, h.logger)
	h.logger.InfoCtx(r.Context(), "order loaded")
}
