package bootstrap

import (
	"context"
	"encoding/hex"
	"envmn/config"
	"envmn/internal/api"
	"envmn/internal/api/grpc"
	grpcauth "envmn/internal/api/grpc/auth"
	envsvc "envmn/internal/domain/environment/services"
	"envmn/internal/domain/environment/services/access"
	domainports "envmn/internal/domain/environment/services/ports"
	"envmn/internal/domain/event"
	infra "envmn/internal/infrastructure"
	metrics "envmn/internal/metircs"
	repo "envmn/internal/repository"
	"envmn/internal/service"
	"envmn/internal/service/ports"
	"envmn/logs"
	"envmn/pkg/retry"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	repositoryCacheRedisDB     = 0
	notifiersRedisDB           = 1
	repositoryCacheMetricsName = "cache.db"
)

var ApiSet = wire.NewSet(
	api.ProvideGRPCServer,
)

var ServiceSet = wire.NewSet(
	service.ProvideDistributionServices,
	service.ProvideManagementServices,
)

var InfraSet = wire.NewSet(
	infra.ProvideReservedEnvironmentsStorage,
	infra.ProvideClientKeyGenerator,
	infra.ProvideKeyGenerator,
	infra.ProvideNotifier,
)

var RepoSet = wire.NewSet(
	repo.ProvideRedisCacheSettings,
	repo.ProvideEnvironmentRepositories,
	repo.ProvideAccessPolicyRepositories,
	repo.ProvideCachedAccessPolicyRepository,
	repo.ProvideCachedAccessPolicyFinderSaver,
	repo.ProvideCachedEnvironmentRepository,
	repo.ProvideCachedEnvironmentPoliciesRepository,
	repo.ProvideCachedEnvironmentVariablesRepository,
)

var MetricsSet = wire.NewSet(
	metrics.ProvideRepositoryCacheMetrics,
)

func providePostgresDB(cfg config.PostgresDBConfig) (*pgxpool.Pool, error) {
	pgConf, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("Failed to parse database config: %w", err)
	}
	pgConf.MaxConns = cfg.MAXConnections
	pgConf.MinConns = cfg.MINConnections

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to create database pool: %w", err)
	}
	return pool, nil
}

type notifiersRedis *redis.Client

func provideNotifiersRedis(cfg config.RedisConfig) (notifiersRedis, error) {
	return redisClient(cfg.Host, cfg.Port, notifiersRedisDB)
}

func provideRepositoryCacheRedis(cfg config.RedisConfig) (repo.RepositoryCacheRedis, error) {
	return redisClient(cfg.Host, cfg.Port, repositoryCacheRedisDB)
}

func provideRootLogger() logs.Logger {
	return logs.NewSlogAdapter(
		slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	)
}

func provideCacheLogs(root logs.Logger) repo.CacheLogger {
	return repo.CacheLogger(root.Child("cache"))
}

type notifiersRetrierLogger logs.Logger

func provideNotifiersRetrierLogger(root logs.Logger) notifiersRetrierLogger {
	return notifiersRetrierLogger(root.Child("notifiers_retrier"))
}

func provideNotifiersRetrier(log notifiersRetrierLogger, cfg config.NotifiersConfig) *retry.Retry {
	return retry.NewRetry(logs.Logger(log), cfg.RetryTimeout, cfg.MaxRetries)
}

func provideKeyGenSeed(cfg config.SecretKeysConfig) (infra.KeySeed, error) {
	key, err := hex.DecodeString(cfg.SeedKey)
	if err != nil {
		return infra.KeySeed{}, err
	}
	return infra.KeySeed(key), nil
}

func provideCacheTTL(cfg config.CacheConfig) repo.CacheTTL {
	return repo.CacheTTL(cfg.CacheTTL)
}

func provideMetricsName() metrics.RepositoryCacheMetricsName {
	return repositoryCacheMetricsName
}

func providePublisher() *event.Publisher {
	return event.NewPublisher()
}

func provideAccessControlService(
	repo domainports.AccessPolicyFinderSaver,
	keyGen domainports.KeyGenerator,
) envsvc.AccessControl {
	return access.New(repo, keyGen)
}

func provideManagmentDependincies(
	envRepo ports.EnvironmentRepository,
	envPolicyRepo ports.EnvironmentPoliciesRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	policyRepo ports.AccessPolicyRepository,
	reservedStorage ports.ReservedEnvironmentsStorage,
	accessControl envsvc.AccessControl,
	publisher *event.Publisher,
) service.ManagmentDependincies {
	return service.ManagmentDependincies{
		EnvironmentRepository:          envRepo,
		ReservedEnvironmentsStorage:    reservedStorage,
		AccessPolicyRepository:         policyRepo,
		EnvironmentPoliciesRepository:  envPolicyRepo,
		EnvironmentVariablesRepository: envVarsRepo,
		Publisher:                      publisher,
		AccessControl:                  accessControl,
	}
}

func provideClientDependincies(
	envRepo ports.EnvironmentRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	reservedStorage ports.ReservedEnvironmentsStorage,
	publisher *event.Publisher,
	accessControl envsvc.AccessControl,
	notifier event.Notifier,
	keyGen ports.ClientKeyGenerator,
) service.DistributionDependincies {
	return service.DistributionDependincies{
		EnvironmentRepository:          envRepo,
		ReservedEnvironmentsStorage:    reservedStorage,
		EnvironmentVariablesRepository: envVarsRepo,
		Publisher:                      publisher,
		AccessControl:                  accessControl,
		Notifier:                       notifier,
		ClientKeyGenerator:             keyGen,
	}
}

func provideNotifierSettings(
	rdb notifiersRedis,
	log logs.Logger,
	retrier *retry.Retry,
) infra.NotifierSettings {
	return infra.NotifierSettings{
		Redis: rdb,
		Log:   log,
		Retry: retrier,
	}
}

func provideGRPCServerSettings(logger logs.Logger, cfg config.CertificateConfig) (grpc.Settigns, error) {
	creds, err := grpcauth.NewMTLSCredentials(cfg.CertPath, cfg.KeyPath, cfg.CACertPath)
	if err != nil {
		return grpc.Settigns{}, err
	}
	return grpc.Settigns{
		Logger:             logger,
		Credentials:        creds,
		PasswordEnvVarName: string(config.PasswordVariableName),
	}, nil
}
