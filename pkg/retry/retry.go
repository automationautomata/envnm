package retry

import (
	"context"
	"envmn/logs"
	"fmt"
	"time"
)

type Retry struct {
	log        logs.Logger
	timeout    time.Duration
	maxRetries int
}

func NewRetry(log logs.Logger, timeout time.Duration, maxRetries int) *Retry {
	return &Retry{
		log:        log,
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

func (retry *Retry) Start(name string, isSync bool, fn func() error) {
	retry.StartWithContext(context.Background(), name, isSync, fn)
}

func (retry *Retry) StartWithContext(ctx context.Context, name string, isSync bool, fn func() error) {
	if isSync {
		retry.start(ctx, name, fn)
		return
	}
	go retry.start(ctx, name, fn)
}

func (retry *Retry) start(ctx context.Context, name string, fn func() error) {
	retry.log.Info(fmt.Sprintf("start retries of %q", name), nil)

	for i := range retry.maxRetries {
		select {
		case <-ctx.Done():
			retry.log.Error(
				"retring finished after context done",
				logs.Args{"name": name, "attemt": i, "error": ctx.Err()},
			)
			return
		case <-time.After(retry.timeout):
			err := fn()
			if err == nil {
				return
			}
			retry.log.Error(
				"retry failed",
				logs.Args{"name": name, "attemt": i, "error": err},
			)
		}
	}
	retry.log.Info("all retries is failed", logs.Args{"name": name})
}
