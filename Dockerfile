FROM docker.io/oven/bun:1-alpine AS frontend-builder

WORKDIR /frontend-build

COPY frontend/package.json frontend/bun.lock ./

RUN bun install --frozen-lockfile

COPY frontend/ ./

RUN bun build src/main.js --outdir /dist --minify && \
    bun build src/main.css --outdir /dist --asset-naming '[name]-[hash].[ext]' --minify

FROM docker.io/library/golang:1.26.3-alpine3.23 AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .

COPY --from=frontend-builder /dist/ ./internal/embed/static/dist/

RUN go tool templ generate
RUN go tool sqlc generate
RUN go tool swag init -g internal/app/app.go --output docs

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath \
    -ldflags="-w -s" \
    -o /build/sosilol \
    ./cmd/sosilol

FROM docker.io/library/alpine:3.23

WORKDIR /app

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup -h /app

COPY --from=builder --chown=appuser:appgroup /build/sosilol .

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:8080/health || exit 1

CMD ["/app/sosilol"]
