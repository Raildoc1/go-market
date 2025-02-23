package timeutils

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAllAttemptsFailed = errors.New("all attempts failed")
)

func Retry[T any](
	ctx context.Context,
	attemptDelays []time.Duration,
	function func(context.Context) (T, error),
	onFinished func(T, error) (needRetry bool),
) (T, error) {
	for _, delay := range attemptDelays {
		if ctx.Err() != nil {
			var res T
			return res, fmt.Errorf("retry canceled: %w", ctx.Err())
		}
		res, err := function(ctx)
		if !onFinished(res, err) {
			return res, err
		}
		err = SleepCtx(ctx, delay)
		if err != nil {
			var res T
			return res, err
		}
	}
	var res T
	return res, ErrAllAttemptsFailed
}

func SleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep canceled: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}
