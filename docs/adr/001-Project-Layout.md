---
name: 001-Project-Layout
title: Project Layout
created: 2026-11-16
pr: https://github.com/sdedovic/wgsltoy/pull/<PR>
depends_on: []
supersedes: []
---

# ADR 001: Project Layout

## 1. Context

### 1.1 Background

The code under `src/go/` has grown without clear boundaries:

- Business rules are tangled with infrastructure. Username and shader validation sit next to use-case steps in `service/`, and the repository turns Postgres errors into business errors.
- One `models` struct serves as the JSON body, the database row, and the domain object, so a change to any one of them touches all three.
- Testing is awkward. A single `IRepository` covers every entity, so each mock embeds all of it, and `goioc/di` wires dependencies through reflection and struct tags, which hides them.
- The request context is thrown away, so cancellation and the authenticated user can't reach the database layer.
- IDs are plain strings, so a user ID can be passed where a shader ID is expected.

### 1.2 Glossary

Terms follow [Domain-Driven Design](https://www.domainlanguage.com/ddd/) and [hexagonal architecture](https://alistair.cockburn.us/hexagonal-architecture/).

## 2. Proposal

### 2.1 Layers

```
cmd/wgsltoy-server/   # entry point
internal/
  app/                # config, logger, fx wiring
  domain/             # entities, value objects, services, ports
  persistence/        # Postgres repositories
  security/           # argon2 and JWT
  server/             # HTTP server and middleware
  controllers/        # HTTP routes and handlers
migrations/
docs/api/             # generated OpenAPI spec
justfile
```

Infrastructure depends on the domain, never the reverse. That keeps the business rules in one place, testable with no database or HTTP, and lets storage or transport change without touching them.

`domain` is one flat package, with files and types prefixed by entity (`user_service.go`, `domain.UserService`). With two entities, sub-packages would add a shared-kernel package and little else. They can be split out later if the domain grows.

### 2.2 Domain types

- Rules live in value-object constructors such as `NewUsername` and `NewShaderName`, so an invalid value can't exist past the boundary.
- IDs are typed (`UserID`, `ShaderID`), so the compiler catches swapped IDs. `NewShaderID(s)` validates and `GenerateShaderID()` creates. Each lives in its entity's file.
- Each layer has its own types: DTOs in controllers, rows in persistence, entities in the domain. Each can change without touching the others.

### 2.3 HTTP

Each controller has a `Register(mux)` method that lists all of its routes, using Go 1.22 method patterns, so a resource's whole API is visible in one place. Each route has exactly one handler method.

All routes move under `/api/v1/` to leave room for a breaking v2. Nothing consumes the API yet, so the old routes are dropped.

### 2.4 Wiring and logging

`go.uber.org/fx` replaces `goioc/di`. Constructors take explicit dependencies, which makes them visible and easy to fake in tests. fx's lifecycle hooks give the database pool and HTTP server a clean start and shutdown.

`go.uber.org/zap` replaces `log` for structured logs, and it plugs into fx's event logger.

### 2.5 Generated code and tooling

- Go 1.27.
- Mocks: mockgen, via a `//go:generate` header in each file that declares interfaces. Mocks are generated into the same package.
- OpenAPI 3.1: swag v2 annotations on controller route methods, generated into `docs/api/`.
- Static checks: go-compiles, then nilaway.
- All of the above are `go.mod` tools, run through `just`: `generate`, `openapi`, `test-unit`, `test-static`, and `test`, which runs every `test-*` target in order.

### 2.6 Open questions

- argon2 and JWT start in `security`. Should any of it live in `domain`?
