dev_database_url := 'postgres://postgres:postgres@localhost:5432/default?sslmode=disable'

# list recipes
default:
    @just --list

# install frontend dependencies
[group('setup')]
[working-directory('web')]
web-install:
    npm ci

# install dev tools outside of the nix shell
[group('setup')]
[working-directory('server')]
tools:
    go install tool
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# start the local postgres container
[group('db')]
db:
    ./scripts/start-local-pg.sh

# apply migrations to the local database
[group('db')]
[working-directory('server')]
migrate:
    migrate -path migrations -database '{{ dev_database_url }}' up

# regenerate mocks
[group('codegen')]
[working-directory('server')]
generate:
    go generate ./...

# regenerate the OpenAPI spec from the controller annotations
[group('codegen')]
[working-directory('server')]
openapi:
    go tool swag init --v3.1 --quiet -g main.go -d cmd/wgsltoy-server,internal/controllers,internal/server -o ../docs/api --ot json,yaml

# run the server and Vite with hot reload; the server restarts on Go and config changes
[group('dev')]
[working-directory('server')]
dev:
    #!/usr/bin/env bash
    set -euo pipefail
    trap 'kill 0' EXIT
    mkdir -p ../tmp
    npm --prefix ../web run dev &
    APP_PROFILE=dev watchexec --restart --exts go,yaml --watch cmd --watch internal \
      --watch application.yaml --watch application-dev.yaml -- \
      'go build -o ../tmp/wgsltoy-server ./cmd/wgsltoy-server && {
        (for _ in $(seq 100); do curl -sf localhost:8080/api/v1/health >/dev/null && touch ../tmp/reload && break; sleep 0.1; done) &
        exec ../tmp/wgsltoy-server
      }'

# build the server and frontend
[group('build')]
build: server-build web-build

# build the server binary into bin/
[group('build')]
[working-directory('server')]
server-build:
    go build -o ../bin/wgsltoy-server ./cmd/wgsltoy-server

# build the frontend into web/dist/
[group('build')]
[working-directory('web')]
web-build:
    npm run build

# build the server wired to the frontend, with nix
[group('build')]
bundle:
    nix build .#bundle

# run all tests
[group('test')]
test: server-test web-test

# run unit tests, static checks, and the linter on the server
[group('test')]
server-test: server-unit server-static server-lint

# run unit tests
[group('test')]
[working-directory('server')]
server-unit:
    go test ./...

# check that everything compiles, then check for nil dereferences
[group('test')]
[working-directory('server')]
server-static:
    go tool go-compiles ./...
    go tool nilaway -include-pkgs github.com/sdedovic/wgsltoy-server -exclude-file-docstrings "Code generated" ./...

# run linter
[group('test')]
[working-directory('server')]
server-lint:
    golangci-lint run ./...

# type check, lint, unit test, and check formatting of the frontend
[group('test')]
[working-directory('web')]
web-test:
    npm run typecheck
    npm run lint
    npm test
    npm run format:check
