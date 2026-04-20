package bootstrap

import (
	"context"
	"encoding/hex"
	"envmn/config"
	"envmn/internal/api/grpc"
	grpcauth "envmn/internal/api/grpc/auth"
	"envmn/internal/domain/environment/services"
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

var InfraSet = wire.NewSet(
	infra.ProvideReservedEnvironmentsStorage,
	infra.ProvideClientKeyGenerator,
	infra.ProvideKeyGenerator,
	infra.ProvideNotifier,
)

var ServiceSet = wire.NewSet(
	service.ProvideClient,
	service.ProvideManagement,
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

func provideEventPublisher() *event.EventPublisher {
	return event.NewEventPublisher()
}

func provideAccessControlService(
	repo services.AccessPolicyFinderSaver,
	keyGen services.KeyGenerator,
) *services.AccessControlService {
	return services.NewAccessControlService(repo, keyGen)
}

func provideManagmentDependincies(
	envRepo ports.EnvironmentRepository,
	envPolicyRepo ports.EnvironmentPoliciesRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	policyRepo ports.AccessPolicyRepository,
	reservedStorage ports.ReservedEnvironmentsStorage,
	accessControl *services.AccessControlService,
	publisher *event.EventPublisher,
) service.ManagmentDependincies {
	return service.ManagmentDependincies{
		EnvironmentRepository:          envRepo,
		ReservedEnvironmentsStorage:    reservedStorage,
		AccessPolicyRepository:         policyRepo,
		EnvironmentPoliciesRepository:  envPolicyRepo,
		EnvironmentVariablesRepository: envVarsRepo,
		EventPublisher:                 publisher,
		AccessControlService:           accessControl,
	}
}

func provideClientDependincies(
	envRepo ports.EnvironmentRepository,
	envVarsRepo ports.EnvironmentVariablesRepository,
	reservedStorage ports.ReservedEnvironmentsStorage,
	publisher *event.EventPublisher,
	accessControl *services.AccessControlService,
	notifier event.Notifier,
	keyGen ports.ClientKeyGenerator,
) service.ClientDependincies {
	return service.ClientDependincies{
		EnvironmentRepository:          envRepo,
		ReservedEnvironmentsStorage:    reservedStorage,
		EnvironmentVariablesRepository: envVarsRepo,
		EventPublisher:                 publisher,
		AccessControlService:           accessControl,
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

func provideServerSettings(logger logs.Logger, cfg config.CertificateConfig) (grpc.ServerSettigns, error) {
	creds, err := grpcauth.NewMTLSCredentials(cfg.CertPath, cfg.KeyPath, cfg.CACertPath)
	if err != nil {
		return grpc.ServerSettigns{}, err
	}
	return grpc.ServerSettigns{
		Logger:      logger,
		Credentials: creds,
	}, nil
}

func provideGrpcServer(
	client *service.Client,
	management *service.Management,
	settings grpc.ServerSettigns,
) (*grpc.Server, error) {
	return grpc.NewServer(client, management, settings)
}
