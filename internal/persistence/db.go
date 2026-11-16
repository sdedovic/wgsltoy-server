//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package persistence

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// keeps a slow query from holding a pooled connection for the whole request
const queryTimeout = 5 * time.Second

const uniqueViolation = "23505"

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Satisfied by both the pool and a transaction, so repositories don't care which they get.
type PGX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Resolved per call rather than injected, so repositories join the transaction in ctx.
type PGXProvider interface {
	GetPGX(ctx context.Context) PGX
}
