package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/jwtauth/v5"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
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

func userIdFromCtx(ctx context.Context) (userId int, err error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	return strconv.Atoi(claims[service.UserIdClaimName].(string))
}

func tryWriteResponseJSON(w http.ResponseWriter, responseItem any) error {
	res, err := json.Marshal(responseItem)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		return err
	}
	return nil
}
