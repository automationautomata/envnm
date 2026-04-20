package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type LogLevel string

const (
	Info  LogLevel = "INFO"
	Debug LogLevel = "DEBUG"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

type CertificateConfig struct {
	CACertPath string `envPrefix:"CA_CERT_PATH"`
	CertPath   string `env:"CERT_PATH"`
	KeyPath    string `env:"KEY_PATH"`
}

type ServerConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
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

type StartupConfig struct {
	Cache       CacheConfig
	Redis       RedisConfig
	Server      ServerConfig      `envPrefix:"ENVMN_"`
	Certificate CertificateConfig `envPrefix:"SERVER_"`
	Notifiers   NotifiersConfig
	DB          PostgresDBConfig
	Keys        SecretKeysConfig
	LogLevel    LogLevel `env:"LOG_LEVEL"`
}

type CLIClientConfig struct {
	Server      ServerConfig      `envPrefix:"ENVMN_"`
	Certificate CertificateConfig `envPrefix:"ENVMN_"`
}

func Load[T StartupConfig | CLIClientConfig]() (T, error) {
	return env.ParseAs[T]()
}

func LoadFromEnvFile[T StartupConfig | CLIClientConfig](path string) (T, error) {
	err := godotenv.Load(path)
	if err != nil {
		var t T
		return t, err
	}
	return Load[T]()
}
