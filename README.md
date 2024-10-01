# WGSL Toy (Server)

[https://wgsltoy.com](https://wgsltoy.com)

## Developing
All dependencies are managed with Nix flake, [flake.nix](./flake.nix).

```
nix develop
```

### Local PostgreSQL
On systems with `docker`, simply run `./scripts/start-local-pg.sh`.

Then apply migrations with:
```bash
migrate -path db/migrations -database 'postgres://postgres:postgres@localhost:5432/default?sslmode=disable' up
```

## PostgreSQL Migrations
Migrations are stored in [`db/migrations`](./db/migrations/) and are executed manually using [go-migrate](https://github.com/golang-migrate/migrate). 
