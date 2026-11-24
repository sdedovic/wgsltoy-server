package persistence

import (
	"context"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

type txManager struct {
	client PgClient
}

var _ domain.Transactor = (*txManager)(nil)

func NewTxManager(client PgClient) domain.Transactor {
	return &txManager{client}
}

// Commits only if fn succeeds. Nested calls join the outer transaction.
func (m *txManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := extractTx(ctx); ok {
		return fn(ctx)
	}

	tx, err := m.client.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(withTx(ctx, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
