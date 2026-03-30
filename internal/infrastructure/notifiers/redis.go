package notifiers

import (
	"context"
	"envmn/internal/domain/event"
	"envmn/internal/service/ports"
	"envmn/logs"
	"envmn/pkg/retry"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const eventsChannelBaseName = "event"

type redisNotifier struct {
	rdb   *redis.Client
	log   logs.Logger
	retry *retry.Retry
}

func NewRedisNotifier(
	rdb *redis.Client,
	log logs.Logger,
	retry *retry.Retry,
) *redisNotifier {
	return &redisNotifier{
		rdb:   rdb,
		retry: retry,
		log:   log,
	}
}

func (ntf *redisNotifier) Notify(ctx context.Context, event event.Event) {
	envName := ctx.Value(ports.EnvironmentNameContextKey).(string)
	chanName := fmt.Sprintf("%s.%s.%s", eventsChannelBaseName, event.Name(), envName)

	payload := event.Payload()
	message, err := payload.MarshalJSON()
	if err != nil {
		ntf.log.Error(
			"cannot marshal payload for event",
			logs.Args{"event_name": event.Name(), "error": err},
		)
		return
	}

	err = ntf.rdb.Publish(ctx, chanName, message).Err()
	if err != nil {
		ntf.log.Error(
			"cannot publish to channel",
			logs.Args{"channel_name": chanName, "error": err},
		)

		name := fmt.Sprintf("%s:%s", chanName, event.OccurredAt())

		ntf.retry.Start(name, false, func() error {
			err := ntf.rdb.Publish(context.Background(), chanName, message).Err()
			if err != nil {
				return fmt.Errorf("cannot publish to channel %q: %w", chanName, err)
			}
			return nil
		})
	}
}
