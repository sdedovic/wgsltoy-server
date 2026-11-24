//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

const queryTimeout = 5 * time.Second

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type PgClient interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type contextKey string

const txKey contextKey = "tx"

func NewPgClient(lc fx.Lifecycle, databaseURL string) (PgClient, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed creating database pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func getClient(ctx context.Context, defaultClient PgClient) PgClient {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return defaultClient
}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func extractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
