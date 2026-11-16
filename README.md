# WGSL Toy (Server)

[https://wgsltoy.com](https://wgsltoy.com)

## Operating
### Configuration
#### Environment Variables
| name           | example                                                               | description                                                  |
|----------------|-----------------------------------------------------------------------|--------------------------------------------------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/default?sslmode=disable` | Postgres connection URL, for storing application data, state |
| `APP_SECRET`   | `test`                                                                | Secret phrase used for signing JWTs for user authentication. Required. |
| `PORT`         | `8080`                                                                | Port the HTTP server listens on. Defaults to `8080`.         |
| `APP_ENV`      | `dev`                                                                 | Set to `dev` for human-readable logs.                        |

## Developing
All dependencies are managed with Nix flake, [flake.nix](./flake.nix).

```
nix develop
```

Common tasks are in the [`justfile`](./justfile):

```
just generate     # regenerate mocks
just openapi      # regenerate the OpenAPI spec in docs/api/
just test         # run all tests
```

Run the server with:
```bash
go run ./cmd/wgsltoy-server
```

The layout is described in [ADR 001](docs/adr/001-Project-Layout.md).

### Database
#### Running Local
On systems with `docker`, simply run `./scripts/start-local-pg.sh`.

Then apply migrations with:
```bash
migrate -path migrations -database 'postgres://postgres:postgres@localhost:5432/default?sslmode=disable' up
```

### Apply Database Migrations
Migrations are stored in [`migrations`](migrations/) and are executed manually using [go-migrate](https://github.com/golang-migrate/migrate).

### Create New Migration

```
migrate create -ext sql -dir migrations -seq <name>
```
