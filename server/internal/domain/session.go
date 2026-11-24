package domain

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

const sessionDuration = 30 * 24 * time.Hour

type SessionID string

type Session struct {
	ID        SessionID
	UserID    UserID
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(userID UserID) (Session, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return Session{}, err
	}
	now := time.Now()
	return Session{
		ID:        SessionID(base64.URLEncoding.EncodeToString(b)),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionDuration),
	}, nil
}
