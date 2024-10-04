package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"time"
)

const userUniqueEmailConstraint = "unique_email"
const userUniqueUsernameConstraint = "unique_username"

type User struct {
	Id                string    `db:"user_id"`
	Username          string    `db:"username"`
	Email             string    `db:"email"`
	EmailVerification string    `db:"email_verification"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func UserInsert(ctx context.Context, pgPool *pgxpool.Pool, username string, email string, passwordHash string) (string, error) {
	now := time.Now()

	sql, args, err := psql.
		Insert("users").Columns("username", "email", "email_verification", "password", "created_at", "updated_at").
		Values(username, email, "pending", passwordHash, now, now).
		Suffix("RETURNING user_id").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed building sql caused by: %w", err)
	}

	var userId string
	err = pgPool.QueryRow(ctx, sql, args...).Scan(&userId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// uniqueness constraint violation
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == userUniqueEmailConstraint {
					return "", infra.NewValidationError("Email is already taken!")
				}

				if pgErr.ConstraintName == userUniqueUsernameConstraint {
					return "", infra.NewValidationError("Username is already taken!")
				}
			}
		}

		// catchall
		return "", fmt.Errorf("failed inserting user caused by: %w", err)
	}

	return userId, nil
}

func UserGetUserIdPasswordByUsername(ctx context.Context, pgPool *pgxpool.Pool, username string) (string, string, error) {
	sql, args, err := psql.Select("user_id", "password").
		From("users").
		Where("username = ?", username).
		ToSql()
	if err != nil {
		return "", "", fmt.Errorf("failed building sql caused by: %w", err)
	}

	var password string
	var userId string
	err = pgPool.QueryRow(ctx, sql, args...).Scan(&userId, &password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", infra.BadLoginError
		}
		return "", "", fmt.Errorf("failed querying user caused by: %w", err)
	}

	return userId, password, nil
}

func UserFindOneById(ctx context.Context, pgPool *pgxpool.Pool, userId string) (*User, error) {
	sql, args, err := psql.
		Select("user_id", "email", "email_verification", "username", "created_at", "updated_at").
		From("users").
		Where("user_id = ?", userId).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed building sql caused by: %w", err)
	}

	rows, err := pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed querying user caused by: %w", err)
	}
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("failed deserializing database rows caused by: %w", err)
	}

	return user, nil
}
