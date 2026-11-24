//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type Transactor interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
