package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type EnvironmentVariableName string

type LogLevel string

const (
	Info  LogLevel = "INFO"
	Debug LogLevel = "DEBUG"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"

	PasswordVariableName EnvironmentVariableName = "ENVMN_PASSWORD"
)

type CertificateConfig struct {
	CACertPath string `env:"CA_CERT_PATH"`
	CertPath   string `env:"CERT_PATH"`
	KeyPath    string `env:"KEY_PATH"`
}

type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

func (cfg *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST"`
	Port int    `env:"REDIS_PORT"`
}

type CacheConfig struct {
	CacheTTL time.Duration `env:"CACHE_TTL"`
}

type NotifiersConfig struct {
	MaxRetries   int           `env:"MAX_RETRIES"`
	RetryTimeout time.Duration `env:"RETRY_TIMEOUT"`
}

type AuthConfig struct {
	Password string `env:"ENVMN_PASSWORD"`
}

type SecretKeysConfig struct {
	SeedKey string `env:"SEED_KEY"`
}

type PostgresDBConfig struct {
	Host           string `env:"POSTGRES_HOST"`
	Port           int    `env:"POSTGRES_PORT"`
	User           string `env:"POSTGRES_USER"`
	Password       string `env:"POSTGRES_PASSWORD"`
	DB             string `env:"POSTGRES_NAME"`
	MAXConnections int32  `env:"POSTGRES_MAX_CONNECTIONS"`
	MINConnections int32  `env:"POSTGRES_MIN_CONNECTIONS"`
}

func (cfg *PostgresDBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB,
	)
}

// //go:generate envdoc -output ./envdocs/startup-env.md
type StartupConfig struct {
	Cache         CacheConfig
	Redis         RedisConfig
	GRPCServer    ServerConfig      `envPrefix:"ENVMN_"`
	MetricsServer ServerConfig      `envPrefix:"METRICS_SERVER_"`
	Certificate   CertificateConfig `envPrefix:"ENVMN_SERVER_"`
	Notifiers     NotifiersConfig
	DB            PostgresDBConfig
	Keys          SecretKeysConfig
	Auth          AuthConfig
	LogLevel      LogLevel `env:"LOG_LEVEL"`
}

// //go:generate envdoc -output ./envdocs/cli-env.md
type CLIConfig struct {
	Auth        AuthConfig
	Server      ServerConfig      `envPrefix:"ENVMN_"`
	Certificate CertificateConfig `envPrefix:"ENVMN_CLIENT_"`
}

func Load[T any]() (T, error) {
	return env.ParseAs[T]()
}

func LoadFromEnvFile[T StartupConfig | CLIConfig](path string) (T, error) {
	err := godotenv.Load(path)
	if err != nil {
		var t T
		return t, err
	}
	return Load[T]()
}
