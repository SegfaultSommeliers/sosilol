FROM docker.io/oven/bun:1.4.4-alpine@sha256:4cf82f5686b196504fe119db3fcd9185b191f82ece82c40680e597228ec6070a AS frontend-builder

WORKDIR /frontend-build

COPY frontend/package.json frontend/bun.lock ./

RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile

COPY frontend/ ./

RUN bun build src/main.js --outdir /dist --minify && \
    bun build src/main.css --outdir /dist --asset-naming '[name]-[hash].[ext]' --minify

FROM docker.io/library/golang:1.27.4-alpine3.24@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY view/ ./view/
RUN --mount=type=cache,target=/go/pkg/mod \
    go tool templ generate

COPY sqlc.yaml ./
COPY db/ ./db/
RUN --mount=type=cache,target=/go/pkg/mod \
    go tool sqlc generate

COPY internal/ ./internal/
RUN --mount=type=cache,target=/go/pkg/mod \
    go tool swag init -g internal/app/app.go --output docs

COPY ["cmd/", "./cmd/"]
COPY --from=frontend-builder /dist/ ./internal/embed/static/dist/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath \
    -ldflags="-w -s" \
    -o /build/sosilol \
    ./cmd/sosilol

FROM docker.io/library/alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

WORKDIR /app

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup -h /app

COPY --from=builder --chown=appuser:appgroup /build/sosilol .

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:8080/health || exit 1

CMD ["/app/sosilol"]
