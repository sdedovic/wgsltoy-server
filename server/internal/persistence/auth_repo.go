package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

type authRepository struct {
	db PgClient
}

var _ domain.AuthRepository = (*authRepository)(nil)

func NewAuthRepository(db PgClient) domain.AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) FindPrincipal(ctx context.Context, id domain.UserID) (domain.Principal, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.Select(userColumns...).From("users").
		Where(squirrel.Eq{"user_id": string(id)}).
		ToSql()
	if err != nil {
		return domain.Principal{}, fmt.Errorf("failed building sql: %w", err)
	}

	rows, err := getClient(ctx, r.db).Query(ctx, sql, args...)
	if err != nil {
		return domain.Principal{}, fmt.Errorf("failed querying principal: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Principal{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Principal{}, fmt.Errorf("failed querying principal: %w", err)
	}

	return domain.Principal{User: row.toDomain()}, nil
}
