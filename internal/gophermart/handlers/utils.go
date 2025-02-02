package handlers

import (
	"context"
	"encoding/json"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"io"
)

func closeBody(ctx context.Context, body io.ReadCloser, logger *logging.ZapLogger) {
	err := body.Close()
	if err != nil {
		logger.ErrorCtx(ctx, "failed to close body", zap.Error(err))
	}
}

func decodeJSON[T any](r io.Reader) (T, error) {
	var out T
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&out)
	return out, err
}
