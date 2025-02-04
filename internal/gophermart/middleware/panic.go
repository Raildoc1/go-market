package middleware

import (
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
)

type PanicRecover struct {
	logger *logging.ZapLogger
}

func NewPanicRecover(logger *logging.ZapLogger) *PanicRecover {
	return &PanicRecover{
		logger: logger,
	}
}

func (pr *PanicRecover) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				pr.logger.ErrorCtx(r.Context(), "panic in HTTP handler", zap.Any("recover", rcv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
