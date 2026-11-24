//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

type PasswordHasher interface {
	Hash(password Password) (string, error)
	Verify(password Password, hash string) (bool, error)
}
