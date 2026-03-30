package infrastructure

import (
	domain "envmn/internal/domain/environment/services"
	"envmn/internal/domain/event"
	"envmn/internal/infrastructure/notifiers"
	"envmn/internal/infrastructure/storages"
	"envmn/internal/service/ports"
	"envmn/logs"
	"envmn/pkg/retry"

	"github.com/redis/go-redis/v9"
)

type KeySeed [32]byte

type NotifierSettings struct {
	Redis *redis.Client
	Log   logs.Logger
	Retry *retry.Retry
}

func ProvideNotifier(settings NotifierSettings) event.Notifier {
	return notifiers.NewRedisNotifier(settings.Redis, settings.Log, settings.Retry)
}

func ProvideReservedEnvironmentsStorage() ports.ReservedEnvironmentsStorage {
	return storages.NewInmemoryEnvironmentsStorage()
}

func ProvideClientKeyGenerator(seed KeySeed) ports.ClientKeyGenerator {
	return NewGenerator(seed)
}

func ProvideKeyGenerator(seed KeySeed) domain.KeyGenerator {
	return NewGenerator(seed)
}
