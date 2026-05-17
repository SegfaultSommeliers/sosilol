package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string `env:"ENVIRONMENT"`

	HttpAddress     string        `env:"HTTP_ADDRESS"`
	GracefulTimeout time.Duration `env:"GRACEFUL_TIMEOUT" envDefault:"10s"`

	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUsername string `env:"POSTGRES_USERNAME" envDefault:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	PostgresDatabase string `env:"POSTGRES_DATABASE" envDefault:"postgres"`

	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	SessionSecret string `env:"SESSION_SECRET"`

	GithubClientId     string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `env:"GITHUB_SECRET"`
	GithubRedirectUrl  string `env:"GITHUB_REDIRECT_URL"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	return env.ParseAs[Config]()
}
