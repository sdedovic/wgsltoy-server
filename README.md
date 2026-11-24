# WGSL Toy (Server)

[https://wgsltoy.com](https://wgsltoy.com)

## Operating
### Configuration
Config is read from [`application.yaml`](application.yaml), then `application-<profile>.yaml` when `APP_PROFILE` is set, then environment variables. Every key in `application.yaml` can be overridden by its upper-cased, underscored name, e.g. `database.url` by `DATABASE_URL`.

| key               | env               | description                                                        |
|-------------------|-------------------|--------------------------------------------------------------------|
| `database.url`    | `DATABASE_URL`    | Postgres connection URL, for storing application data, state. Required. |
| `app.secret`      | `APP_SECRET`      | Secret phrase used for signing user sessions. Required.            |
| `server.port`     | `SERVER_PORT`     | Port the HTTP server listens on. Defaults to `8080`.               |
| `log.development` | `LOG_DEVELOPMENT` | `true` for human-readable logs, `false` for JSON. Defaults to `false`. |
| `templates.dir`   | `TEMPLATES_DIR`   | Directory of HTML templates. Defaults to `templates`.              |

## Developing
All dependencies are managed with Nix flake, [flake.nix](./flake.nix).

```
nix develop
```

Common tasks are in the [`justfile`](./justfile):

```
just tools        # install dev tools into GOBIN, if not using nix
just generate     # regenerate mocks
just openapi      # regenerate the OpenAPI spec in docs/api/
just test         # run all tests
```

Run the server with:
```bash
go run ./cmd/wgsltoy-server
```

Set `APP_PROFILE=dev` to load [`application-dev.yaml`](application-dev.yaml), which points at the local database below.

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
