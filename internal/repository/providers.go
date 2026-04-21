package repository

import (
	"envmn/internal/repository/cache"
	"envmn/internal/repository/decorators"
	"envmn/internal/repository/postgres"
	"envmn/internal/service/ports"

	domainports "envmn/internal/domain/environment/services/ports"
	"envmn/logs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type EnvironmentRepositories struct {
	ports.EnvironmentRepository
	ports.EnvironmentPoliciesRepository
	ports.EnvironmentVariablesRepository
}

type AccessPolicyRepositories struct {
	domainports.AccessPolicyFinderSaver
	ports.AccessPolicyRepository
}

func ProvideEnvironmentRepositories(db *pgxpool.Pool) EnvironmentRepositories {
	return EnvironmentRepositories{
		postgres.NewEnvironmentsRepository(postgres.NewConnection(db)),
		postgres.NewAccessPoliciesRepository(postgres.NewConnection(db)),
		postgres.NewVariablesRepository(postgres.NewConnection(db)),
	}
}

func ProvideAccessPolicyRepositories(db *pgxpool.Pool) AccessPolicyRepositories {
	return AccessPolicyRepositories{
		postgres.NewAccessPoliciesRepository(postgres.NewConnection(db)),
		postgres.NewAccessPoliciesRepository(postgres.NewConnection(db)),
	}
}

type (
	RepositoryCacheRedis *redis.Client
	CacheLogger          logs.Logger
	CacheTTL             time.Duration
)

func ProvideRedisCacheSettings(
	rdb RepositoryCacheRedis,
	ttl CacheTTL,
	metrics cache.RepositoryMetrics,
) cache.RedisCacheSettings {
	return cache.RedisCacheSettings{Redis: rdb, TTL: time.Duration(ttl)}
}

func ProvideCachedEnvironmentRepository(
	log CacheLogger,
	settings cache.RedisCacheSettings,
	repos EnvironmentRepositories,
) ports.EnvironmentRepository {
	return decorators.NewEnvironmentRepositoryCache(
		log,
		settings,
		repos.EnvironmentRepository,
		repos.EnvironmentVariablesRepository,
		repos.EnvironmentPoliciesRepository,
	)
}

func ProvideCachedEnvironmentVariablesRepository(
	log CacheLogger,
	settings cache.RedisCacheSettings,
	repos EnvironmentRepositories,
) ports.EnvironmentVariablesRepository {
	return decorators.NewEnvironmentRepositoryCache(
		log,
		settings,
		repos.EnvironmentRepository,
		repos.EnvironmentVariablesRepository,
		repos.EnvironmentPoliciesRepository,
	)
}

func ProvideCachedEnvironmentPoliciesRepository(
	log CacheLogger,
	settings cache.RedisCacheSettings,
	repos EnvironmentRepositories,
) ports.EnvironmentPoliciesRepository {
	return decorators.NewEnvironmentRepositoryCache(
		log,
		settings,
		repos.EnvironmentRepository,
		repos.EnvironmentVariablesRepository,
		repos.EnvironmentPoliciesRepository,
	)
}

func ProvideCachedAccessPolicyFinderSaver(
	log CacheLogger,
	settings cache.RedisCacheSettings,
	repos AccessPolicyRepositories,
) domainports.AccessPolicyFinderSaver {
	return decorators.NewPolicyRepositoryCache(
		log,
		settings,
		repos.AccessPolicyRepository,
		repos.AccessPolicyFinderSaver,
	)
}

func ProvideCachedAccessPolicyRepository(
	log CacheLogger,
	settings cache.RedisCacheSettings,
	repos AccessPolicyRepositories,
) ports.AccessPolicyRepository {
	return decorators.NewPolicyRepositoryCache(
		log,
		settings,
		repos.AccessPolicyRepository,
		repos.AccessPolicyFinderSaver,
	)
}
