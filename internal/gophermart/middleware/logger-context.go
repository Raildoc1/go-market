package middleware

import (
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"net/http"
)

type LoggerContext struct{}

func NewLoggerContext() *LoggerContext {
	return &LoggerContext{}
}

func (lc *LoggerContext) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(
			logging.WithContextFields(
				r.Context(),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote-addr", r.RemoteAddr),
			),
		)
		next.ServeHTTP(w, r)
	})
}
