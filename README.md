# WGSL Toy (Server)

[https://wgsltoy.com](https://wgsltoy.com)

## Operating
### Configuration
Config is read from [`server/application.yaml`](server/application.yaml), then `application-<profile>.yaml` when `APP_PROFILE` is set, then environment variables. Every key in `application.yaml` can be overridden by its upper-cased, underscored name, e.g. `database.url` by `DATABASE_URL`.

| key               | env               | description                                                        |
|-------------------|-------------------|--------------------------------------------------------------------|
| `database.url`    | `DATABASE_URL`    | Postgres connection URL, for storing application data, state. Required. |
| `app.secret`      | `APP_SECRET`      | Secret phrase used for signing user sessions. Required.            |
| `server.port`     | `SERVER_PORT`     | Port the HTTP server listens on. Defaults to `8080`.               |
| `log.development` | `LOG_DEVELOPMENT` | `true` for human-readable logs, `false` for JSON. Defaults to `false`. |
| `templates.dir`   | `TEMPLATES_DIR`   | Directory of HTML templates. Defaults to `../web/templates`, relative to `server/`.          |
| `static.dir`      | `STATIC_DIR`      | Output of `vite build`, served at `/static/`. Defaults to `../web/dist`, relative to `server/`. |
| `vite.url`        | `VITE_URL`        | Vite dev server to load the frontend from instead of `static.dir`. Empty by default. |

## Developing
All dependencies are managed with Nix flake, [flake.nix](./flake.nix).

```
nix develop
```

Common tasks are in the [`justfile`](./justfile):

```
just tools        # install dev tools into GOBIN, if not using nix
just web-install  # install frontend dependencies
just generate     # regenerate mocks
just openapi      # regenerate the OpenAPI spec in docs/api/
just dev          # run the server and frontend with hot reload
just test         # run all tests
```

`just dev` loads [`server/application-dev.yaml`](server/application-dev.yaml), which points at the local database below, and serves the site at http://localhost:8080.

- CSS and React components hot reload in place, and pages reload when a template changes.
- The server rebuilds and restarts on Go and config changes, and open pages reload once it's back.

For production, build the frontend with `just web-build`, then run the server, which serves `web/dist`.

### Frontend
The frontend lives in [`web/`](web/), a self-contained npm package built with [Vite](https://vite.dev).

```
web/
  templates/   # Go html/template pages, rendered by the server per request
  src/         # TypeScript and CSS, built by Vite into dist/
```

- Every page is rendered inside `templates/layout.html`, with the templates in `templates/partials/`.
- Templates load scripts with `{{vite "src/editor.tsx"}}`.
- Pages are static HTML. Interactive parts are React islands: `<div data-island="name" data-*>`, mounted by the page's entry with the data attributes as props.

The layout is described in [ADR 001](docs/adr/001-Project-Layout.md).

### Database
#### Running Local
On systems with `docker`, simply run `./scripts/start-local-pg.sh`.

Then apply migrations with:
```bash
cd server && migrate -path migrations -database 'postgres://postgres:postgres@localhost:5432/default?sslmode=disable' up
```

### Apply Database Migrations
Migrations are stored in [`server/migrations`](server/migrations/) and are executed manually using [go-migrate](https://github.com/golang-migrate/migrate).

### Create New Migration

```
migrate create -ext sql -dir server/migrations -seq <name>
```
