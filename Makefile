.DEFAULT_GOAL := build

.PHONY: build clean

build:
	cd frontend && bun install --frozen-lockfile
	cd frontend && bun run build
	go tool templ generate
	go tool sqlc generate
	go tool swag init -g internal/app/app.go --output docs
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o sosilol ./cmd/sosilol

clean:
	rm -f sosilol
	rm -rf internal/embed/static/dist/
	rm -f view/*_templ.go
	rm -rf internal/db/
	rm -rf docs/
	rm -rf frontend/node_modules/
