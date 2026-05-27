package config

import (
	"errors"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

func (e Environment) IsDev() bool { return e == EnvDev }

type Config struct {
	Environment Environment `env:"ENVIRONMENT" envDefault:"prod"`

	HttpAddress     string        `env:"HTTP_ADDRESS,required"`
	GracefulTimeout time.Duration `env:"GRACEFUL_TIMEOUT" envDefault:"10s"`

	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUsername string `env:"POSTGRES_USERNAME" envDefault:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	PostgresDatabase string `env:"POSTGRES_DATABASE" envDefault:"postgres"`
	PostgresTLS      bool   `env:"POSTGRES_TLS" envDefault:"false"`

	RedisHost     string        `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     string        `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string        `env:"REDIS_PASSWORD" envDefault:""`
	RedisTLS      bool          `env:"REDIS_TLS" envDefault:"false"`
	PasteCacheTTL time.Duration `env:"PASTE_CACHE_TTL" envDefault:"1h"`

	TrustedProxy string `env:"TRUSTED_PROXY" envDefault:""`

	GithubClientId     string `env:"GITHUB_CLIENT_ID,required"`
	GithubClientSecret string `env:"GITHUB_SECRET,required"`
	GithubRedirectUrl  string `env:"GITHUB_REDIRECT_URL,required"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return cfg, err
	}

	if !cfg.Environment.IsDev() {
		u, parseErr := url.Parse(cfg.GithubRedirectUrl)
		if parseErr != nil || u.Scheme != "https" {
			return cfg, errors.New("GITHUB_REDIRECT_URL must use https:// in production")
		}
	}

	return cfg, nil
}
