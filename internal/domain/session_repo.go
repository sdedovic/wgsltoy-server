//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type SessionRepository interface {
	Save(ctx context.Context, session Session) error
	FindByID(ctx context.Context, id SessionID) (Session, error)
	DeleteByID(ctx context.Context, id SessionID) error
	DeleteByUserID(ctx context.Context, userID UserID) error
}
