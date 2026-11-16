package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type TxManager struct {
	pool *pgxpool.Pool
}

var _ PGXProvider = (*TxManager)(nil)

// Connects lazily, so Ping it on startup to fail fast.
func NewTxManager(databaseURL string) (*TxManager, error) {
	// TODO: handle TLS
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing database url caused by: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed creating database pool caused by: %w", err)
	}

	return &TxManager{pool}, nil
}

func (m *TxManager) Ping(ctx context.Context) error {
	return m.pool.Ping(ctx)
}

func (m *TxManager) Close() {
	m.pool.Close()
}

func (m *TxManager) GetPGX(ctx context.Context) PGX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return m.pool
}

// Commits only if fn succeeds. Nested calls join the outer transaction.
func (m *TxManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}
	return pgx.BeginFunc(ctx, m.pool, func(tx pgx.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}
