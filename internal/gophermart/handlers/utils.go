package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func closeBody(body io.ReadCloser, logger *zap.Logger) {
	err := body.Close()
	if err != nil {
		logger.Error("failed to close body", zap.Error(err))
	}
}

func newRequestLogger(logger *zap.Logger, r *http.Request) *zap.Logger {
	return logger.With(
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("remote-addr", r.RemoteAddr),
	)
}

func decodeJSON[T any](r io.Reader) (T, error) {
	var out T
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&out)
	return out, err
}
