//go:build wireinject
// +build wireinject

package bootstrap

import (
	"context"
	"envmn/config"
	"envmn/internal/api/grpc"
	"envmn/internal/cache"
	"envmn/logs"
	"fmt"
	"net"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	grpcServer     *grpc.Server
	db             *pgxpool.Pool
	cacheRedis     *redis.Client
	notifiersRedis *redis.Client
	logger         logs.Logger
	serverConf     config.ServerConfig
}

func newApp(
	grpcServer *grpc.Server,
	db *pgxpool.Pool,
	cacheRedis cache.RepositoryCacheRedis,
	notifiersRedis notifiersRedis,
	logger logs.Logger,
	cfg config.ServerConfig,
) *App {
	return &App{
		grpcServer:     grpcServer,
		db:             db,
		cacheRedis:     (*redis.Client)(cacheRedis),
		notifiersRedis: (*redis.Client)(notifiersRedis),
		logger:         logger,
		serverConf:     cfg,
	}
}

func (a *App) shutdown(ctx context.Context) error {
	a.grpcServer.GracefulStop()
	a.db.Close()
	a.cacheRedis.Close()
	a.notifiersRedis.Close()
	return nil
}

func (a *App) Run() func(ctx context.Context) error {
	go func() {
		addr := fmt.Sprintf("%s:%d", a.serverConf.Host, a.serverConf.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			a.logger.Error("Failed to listen", logs.Args{"error": err})
			return
		}
		if err := a.grpcServer.Serve(lis); err != nil {
			a.logger.Error("Failed to serve", logs.Args{"error": err})
		}
	}()
	return a.shutdown
}

func Build(cfg config.StartupConfig) (*App, error) {
	app, err := InitializeApp(cfg)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Wire injector function
func InitializeApp(cfg config.StartupConfig) (*App, error) {
	wire.Build(
		wire.FieldsOf(&cfg,
			"Server",
			"DB",
			"Redis",
			"Cache",
			"Notifiers",
			"Certificate",
			"Keys",
		),
		provideRootLogger,
		provideCacheLogs,
		provideNotifiersRetrierLogger,
		provideNotifiersRetrier,
		providePostgresDB,
		provideRepositoryCacheRedis,
		provideCacheTTL,
		provideMetricsName,
		provideMetricsNameString,
		provideEventPublisher,
		provideAccessControlService,
		RepoSet,
		InfraSet,
		CacheSet,
		ServiceSet,
		provideServerSettings,
		provideGrpcServer,
		newApp,
	)
	return nil, nil
}
