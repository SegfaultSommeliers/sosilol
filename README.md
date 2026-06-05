# sosi.lol

Amazing paste service with beautiful name.

Url: https://sosi.lol

![Screenshot](https://i.imgur.com/fpMcNLP.png "Screenshot")

## Running

The easiest way — Docker Compose (starts the app, PostgreSQL, Valkey, and runs migrations automatically):

```bash
docker compose up --build -d
```

## Building manually

Requires **Go 1.26.3+** and **Bun**.

```bash
make build
```

To clean all build artifacts and generated files:

```bash
make clean
```

## Environment variables

Copy `.env.example` and fill in the required values:

```bash
cp .env.example .env
```

| Variable | Required | Default | Description |
|---|---|---|---|
| `GITHUB_CLIENT_ID` | **yes** | — | GitHub OAuth App client ID |
| `GITHUB_SECRET` | **yes** | — | GitHub OAuth App client secret |
| `GITHUB_REDIRECT_URL` | **yes** | — | OAuth redirect URL (e.g. `https://yourdomain.com/auth/redirect`) |
| `HTTP_ADDRESS` | **yes** | — | Address to listen on (e.g. `:8080`) |
| `ENVIRONMENT` | no | `prod` | Set to `dev` to disable HTTPS-only cookies and HSTS |
| `GRACEFUL_TIMEOUT` | no | `10s` | Graceful shutdown timeout |
| `POSTGRES_HOST` | no | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | no | `5432` | PostgreSQL port |
| `POSTGRES_USERNAME` | no | `postgres` | PostgreSQL username |
| `POSTGRES_PASSWORD` | no | `postgres` | PostgreSQL password |
| `POSTGRES_DATABASE` | no | `postgres` | PostgreSQL database name |
| `POSTGRES_TLS` | no | `false` | Enable TLS for PostgreSQL connection |
| `REDIS_HOST` | no | `localhost` | Redis/Valkey host |
| `REDIS_PORT` | no | `6379` | Redis/Valkey port |
| `REDIS_PASSWORD` | no | `` | Redis/Valkey password |
| `REDIS_TLS` | no | `false` | Enable TLS for Redis connection |
| `PASTE_CACHE_TTL` | no | `1h` | How long pastes are cached in Redis |
| `TRUSTED_PROXY` | no | `` | IP of your reverse proxy for `X-Forwarded-For` trust |

## TODO

- Localization
- Optimize saving pastes to profiles
- ...

## Bugs and feedback

For bugs, questions and discussions please use the [GitHub Issues](https://github.com/SegfaultSommeliers/sosilol/issues).
