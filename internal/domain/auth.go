//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type Identity struct {
	UserID UserID
}

type identityKey struct{}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, identity)
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey{}).(Identity)
	return identity, ok
}

type PasswordHasher interface {
	Hash(password Password) (string, error)
	Verify(password Password, hash string) (bool, error)
}

type TokenIssuer interface {
	Issue(identity Identity) (string, error)
}

type TokenVerifier interface {
	// ErrUnauthorized for invalid tokens
	Verify(token string) (Identity, error)
}
