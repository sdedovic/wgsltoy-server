# WGSL Toy (Server)

[https://wgsltoy.com](https://wgsltoy.com)

## Operating
### Configuration
#### Environment Variables
| name           | example                                                               | description                                                  |
|----------------|-----------------------------------------------------------------------|--------------------------------------------------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/default?sslmode=disable` | Postgres connection URL, for storing application data, state |
| `APP_SECRET`   | `test`                                                                | Secret phrase used for signing JWTs for user authentication  |

## Developing
All dependencies are managed with Nix flake, [flake.nix](./flake.nix).

```
nix develop
```

### Database
#### Running Local
On systems with `docker`, simply run `./scripts/start-local-pg.sh`.

Then apply migrations with:
```bash
migrate -path src/sql/migrations -database 'postgres://postgres:postgres@localhost:5432/default?sslmode=disable' up
```

### Apply Database Migrations
Migrations are stored in [`db/migrations`](src/sql/migrations/) and are executed manually using [go-migrate](https://github.com/golang-migrate/migrate).

### Create New Migration

```
migrate create -ext sql -dir src/sql/migrations -seq <name>
```