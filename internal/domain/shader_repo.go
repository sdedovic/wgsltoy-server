//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type ShaderRepository interface {
	Save(ctx context.Context, shader Shader) error
	FindByID(ctx context.Context, id ShaderID) (Shader, error)
	FindTop100ByCreatedByOrderByUpdatedAtDesc(ctx context.Context, createdBy UserID) ([]ShaderSummary, error)
}
