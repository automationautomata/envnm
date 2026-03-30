package cache

import (
	metrics "envmn/internal/cache/metircs"
	repocache "envmn/internal/cache/repository"
	"envmn/internal/domain/environment/services"
	"envmn/internal/service/ports"
	"envmn/logs"
	"time"

	"github.com/redis/go-redis/v9"
)

type (
	RepositoryCacheRedis       *redis.Client
	CacheLogger                logs.Logger
	CacheTTL                   time.Duration
	RepositoryCacheMetricsName string
)

func ProvideRedisCacheSettings(rdb RepositoryCacheRedis, log CacheLogger, ttl CacheTTL) repocache.RedisCacheSettings {
	return repocache.RedisCacheSettings{Redis: rdb, Log: log, TTL: time.Duration(ttl)}
}

func ProvideEnvironmentRepositoryCache(
	metricsName RepositoryCacheMetricsName,
	settings repocache.RedisCacheSettings,
	envRepo ports.EnvironmentRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	policiesRepo ports.EnvironmentPoliciesRepository,
) (EnvironmentRepositoryCache, error) {
	m, err := metrics.NewRepositoryMetrics(string(metricsName))
	if err != nil {
		return nil, err
	}
	return repocache.NewEnvironmentCache(settings, envRepo, envVarsRepo, policiesRepo, m), nil
}

func ProvidePolicyRepositoryCache(
	metricsName string,
	settings repocache.RedisCacheSettings,
	policyRepo ports.AccessPolicyRepository,
	finderSaver services.AccessPolicyFinderSaver,
) (PolicyRepositoryCache, error) {
	m, err := metrics.NewRepositoryMetrics(metricsName)
	if err != nil {
		return nil, err
	}
	return repocache.NewPolicyCache(settings, policyRepo, finderSaver, m), nil
}
