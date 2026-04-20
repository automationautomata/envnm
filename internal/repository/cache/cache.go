package cache

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RepositoryMetrics interface {
	Inc(table, field string, hit bool)
}

type RedisCacheSettings struct {
	Redis   *redis.Client
	TTL     time.Duration
	Metrics RepositoryMetrics
}

type redisCache struct {
	rdb     *redis.Client
	ttl     time.Duration
	metrics RepositoryMetrics
}

func newRedisCache(settings RedisCacheSettings) *redisCache {
	return &redisCache{
		rdb:     settings.Redis,
		ttl:     settings.TTL,
		metrics: settings.Metrics,
	}
}

func (cache *redisCache) makeKey(args ...string) string {
	return strings.Join(args, ":")
}

func (cache *redisCache) Hit(table, field string) {
	cache.metrics.Inc(table, field, true)
}

func (cache *redisCache) Miss(table, field string) {
	cache.metrics.Inc(table, field, true)
}
