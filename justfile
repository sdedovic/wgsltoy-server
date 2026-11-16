# list recipes
default:
    @just --list

# regenerate mocks
generate:
    go generate ./...

# regenerate the OpenAPI spec from the controller annotations
openapi:
    go tool swag init --v3.1 --quiet -g main.go -d cmd/wgsltoy-server,internal/controllers,internal/server -o docs/api --ot json,yaml

# run unit tests
test-unit:
    go test ./...

# check that everything compiles, then check for nil dereferences
test-static:
    go tool go-compiles ./...
    go tool nilaway -include-pkgs github.com/sdedovic/wgsltoy-server -exclude-file-docstrings "Code generated" ./...

# run all tests
test: test-unit test-static
