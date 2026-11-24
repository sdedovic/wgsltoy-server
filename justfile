# list recipes
default:
    @just --list

# install dev tools outside of the nix shell
[group('setup')]
tools:
    go install tool
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# regenerate mocks
[group('codegen')]
generate:
    go generate ./...

# regenerate the OpenAPI spec from the controller annotations
[group('codegen')]
openapi:
    go tool swag init --v3.1 --quiet -g main.go -d cmd/wgsltoy-server,internal/controllers,internal/server -o docs/api --ot json,yaml

# run all tests
[group('test')]
test: test-unit test-static lint

# run unit tests
[group('test')]
test-unit:
    go test ./...

# check that everything compiles, then check for nil dereferences
[group('test')]
test-static:
    go tool go-compiles ./...
    go tool nilaway -include-pkgs github.com/sdedovic/wgsltoy-server -exclude-file-docstrings "Code generated" ./...

# run linter
[group('test')]
lint:
    golangci-lint run ./...
