package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

var userColumns = []string{"user_id", "created_at", "updated_at", "username", "email", "email_verification", "password"}

type userRow struct {
	ID                string    `db:"user_id"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
	Username          string    `db:"username"`
	Email             string    `db:"email"`
	EmailVerification string    `db:"email_verification"`
	Password          string    `db:"password"`
}

func (r userRow) toDomain() domain.User {
	return domain.User{
		ID:                domain.UserID(r.ID),
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		Username:          domain.Username(r.Username),
		Email:             domain.Email(r.Email),
		EmailVerification: domain.EmailVerification(r.EmailVerification),
		PasswordHash:      r.Password,
	}
}

type UserRepository struct {
	db PGXProvider
}

var _ domain.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db PGXProvider) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Insert("users").
		Columns(userColumns...).
		Values(string(user.ID), user.CreatedAt, user.UpdatedAt, string(user.Username), string(user.Email), string(user.EmailVerification), user.PasswordHash).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET " +
			"updated_at = EXCLUDED.updated_at, username = EXCLUDED.username, email = EXCLUDED.email, " +
			"email_verification = EXCLUDED.email_verification, password = EXCLUDED.password").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed building sql caused by: %w", err)
	}

	_, err = r.db.GetPGX(ctx).Exec(ctx, sql, args...)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		switch pgErr.ConstraintName {
		case "unique_email":
			return domain.ErrEmailTaken
		case "unique_username":
			return domain.ErrUsernameTaken
		}
	}
	if err != nil {
		return fmt.Errorf("failed saving user caused by: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username domain.Username) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.Select(userColumns...).From("users").Where(squirrel.Eq{"username": string(username)}).ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, err := r.db.GetPGX(ctx).Query(ctx, sql, args...)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	return row.toDomain(), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.Select(userColumns...).From("users").Where(squirrel.Eq{"user_id": string(id)}).ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, err := r.db.GetPGX(ctx).Query(ctx, sql, args...)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("failed querying user caused by: %w", err)
	}

	return row.toDomain(), nil
}
