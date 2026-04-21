//go:generate go run github.com/google/wire/cmd/wire

//go:build wireinject
// +build wireinject

package bootstrap

import (
	"context"
	"envmn/config"
	grpcapi "envmn/internal/api/grpc"
	repo "envmn/internal/repository"
	"envmn/logs"
	"fmt"
	"net"
	"net/http"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type (
	httpServerConfig config.ServerConfig
	gRPCServerConfig config.ServerConfig
)

type App struct {
	grpcServer     *grpcapi.Server
	httpServer     *http.Server
	db             *pgxpool.Pool
	cacheRedis     *redis.Client
	notifiersRedis *redis.Client
	log            logs.Logger
	grpcServerConf config.ServerConfig
	httpServerConf config.ServerConfig
}

func newApp(
	grpcServer *grpcapi.Server,
	db *pgxpool.Pool,
	cacheRedis repo.RepositoryCacheRedis,
	notifiersRedis notifiersRedis,
	log logs.Logger,
	grpcServerConf gRPCServerConfig,
	httpServerConf httpServerConfig,
) *App {
	return &App{
		grpcServer:     grpcServer,
		db:             db,
		cacheRedis:     (*redis.Client)(cacheRedis),
		notifiersRedis: (*redis.Client)(notifiersRedis),
		log:            log,
		grpcServerConf: config.ServerConfig(grpcServerConf),
		httpServerConf: config.ServerConfig(httpServerConf),
	}
}

func (app *App) shutdown(ctx context.Context) error {
	app.grpcServer.GracefulStop()
	if err := app.httpServer.Shutdown(ctx); err != nil {
		app.log.Error("HTTP server graceful shutdown failed: %v", logs.Args{"error": err})
	}
	app.db.Close()
	app.cacheRedis.Close()
	app.notifiersRedis.Close()
	return nil
}

func (app *App) Run() func(context.Context) error {
	go func() {
		var g errgroup.Group

		g.Go(app.runGRPC)
		g.Go(app.runHTTP)

		if err := g.Wait(); err != nil {
			app.log.Error("Application failed", logs.Args{"error": err})
		}
	}()
	return app.shutdown
}

func (app *App) runGRPC() error {
	addr := app.grpcServerConf.Addr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	if err = app.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("grpc server failed: %w", err)
	}
	return nil
}

func (app *App) runHTTP() error {
	addr := app.httpServerConf.Addr()

	app.httpServer = &http.Server{
		Addr:    addr,
		Handler: promhttp.Handler(),
	}

	fmt.Printf("Server prometheus metrics server on %s...\n", addr)
	if err := app.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
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
		provideKeyGenSeed,
		provideNotifiersRedis,
		provideNotifierSettings,
		providePublisher,
		provideAccessControlService,
		provideClientDependincies,
		provideManagmentDependincies,
		provideGRPCServerSettings,
		provideHTTPServerConf,
		provideGRPCServerConf,
		RepoSet,
		InfraSet,
		ServiceSet,
		MetricsSet,
		ApiSet,
		newApp,
	)
	return nil, nil
}
