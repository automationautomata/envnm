package repository

import (
	"envmn/logs"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type RepositoryMetrics interface {
	Inc(table, field string, hit bool)
}

type RedisCacheSettings struct {
	Redis *redis.Client
	Log   logs.Logger
	TTL   time.Duration
}

type redisCacheDecorator struct {
	rdb     *redis.Client
	log     logs.Logger
	ttl     time.Duration
	metrics RepositoryMetrics
	group   *singleflight.Group
}

func newRedisCacheDecorator(settings RedisCacheSettings, metrics RepositoryMetrics) *redisCacheDecorator {
	return &redisCacheDecorator{
		rdb:     settings.Redis,
		log:     settings.Log,
		ttl:     settings.TTL,
		metrics: metrics,
		group:   &singleflight.Group{},
	}
}

func (cache *redisCacheDecorator) makeKey(args ...string) string {
	return strings.Join(args, ":")
}
