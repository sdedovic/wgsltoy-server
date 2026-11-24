package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

type sessionRepo struct {
	db PgClient
}

func NewSessionRepository(db PgClient) domain.SessionRepository {
	return &sessionRepo{db}
}

func (r *sessionRepo) Save(ctx context.Context, session domain.Session) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Insert("sessions").
		Columns("session_id", "user_id", "created_at", "expires_at").
		Values(string(session.ID), string(session.UserID), session.CreatedAt, session.ExpiresAt).
		Suffix("ON CONFLICT (session_id) DO UPDATE SET expires_at = EXCLUDED.expires_at").
		ToSql()
	if err != nil {
		return err
	}

	_, err = getClient(ctx, r.db).Exec(ctx, sql, args...)
	return err
}

func (r *sessionRepo) FindByID(ctx context.Context, id domain.SessionID) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Select("session_id", "user_id", "created_at", "expires_at").
		From("sessions").
		Where("session_id = ? AND expires_at > ?", string(id), time.Now()).
		ToSql()
	if err != nil {
		return domain.Session{}, err
	}

	rows, err := getClient(ctx, r.db).Query(ctx, sql, args...)
	if err != nil {
		return domain.Session{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return domain.Session{}, err
		}
		return domain.Session{}, domain.ErrNotFound
	}

	var row struct {
		SessionID string
		UserID    string
		CreatedAt time.Time
		ExpiresAt time.Time
	}
	if err := rows.Scan(&row.SessionID, &row.UserID, &row.CreatedAt, &row.ExpiresAt); err != nil {
		return domain.Session{}, err
	}

	return domain.Session{
		ID:        domain.SessionID(row.SessionID),
		UserID:    domain.UserID(row.UserID),
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

func (r *sessionRepo) DeleteByID(ctx context.Context, id domain.SessionID) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Delete("sessions").
		Where("session_id = ?", string(id)).
		ToSql()
	if err != nil {
		return err
	}

	_, err = getClient(ctx, r.db).Exec(ctx, sql, args...)
	return errors.Join(err)
}

func (r *sessionRepo) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	sql, args, err := psql.
		Delete("sessions").
		Where("user_id = ?", string(userID)).
		ToSql()
	if err != nil {
		return err
	}

	_, err = getClient(ctx, r.db).Exec(ctx, sql, args...)
	return errors.Join(err)
}
