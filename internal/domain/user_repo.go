//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type UserRepository interface {
	// ErrUsernameTaken or ErrEmailTaken on conflict
	Save(ctx context.Context, user User) error
	FindByUsername(ctx context.Context, username Username) (User, error)
	FindByID(ctx context.Context, id UserID) (User, error)
}
