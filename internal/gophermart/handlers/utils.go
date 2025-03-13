package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-market/internal/gophermart/service"
	"go-market/pkg/logging"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

const (
	invalidUserID = -1
)

var (
	failedToRecoverUserIDErrorMessage = "failed to recover user id"
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
	return out, err //nolint:wrapcheck // unnecessary
}

func userIDFromCtx(ctx context.Context) (userID int, err error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userIDStr, ok := claims[service.UserIDClaimName].(string)
	if !ok {
		return invalidUserID, errors.New("invalid user id type")
	}
	userID, err = strconv.Atoi(userIDStr)
	if err != nil {
		return invalidUserID, fmt.Errorf("failed to parse user id: %w", err)
	}
	return userID, nil
}

func tryWriteResponseJSON(w http.ResponseWriter, responseItem any) error {
	res, err := json.Marshal(responseItem)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	return nil
}
