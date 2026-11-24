//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import (
	"context"
	"errors"
)

type AuthRepository interface {
	FindPrincipal(ctx context.Context, id UserID) (Principal, error)
}

type AuthService interface {
	Authenticate(ctx context.Context, sessionID string) (Principal, error)
}

type authService struct {
	sessions SessionRepository
	repo     AuthRepository
}

var _ AuthService = (*authService)(nil)

func NewAuthService(sessions SessionRepository, repo AuthRepository) AuthService {
	return &authService{sessions, repo}
}

func (s *authService) Authenticate(ctx context.Context, sessionID string) (Principal, error) {
	session, err := s.sessions.FindByID(ctx, SessionID(sessionID))
	if errors.Is(err, ErrNotFound) {
		return Principal{}, ErrUnauthorized
	}
	if err != nil {
		return Principal{}, err
	}
	return s.repo.FindPrincipal(ctx, session.UserID)
}
