//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package domain

import "context"

type ShaderRepository interface {
	Save(ctx context.Context, shader Shader) error
	FindByID(ctx context.Context, id ShaderID) (Shader, error)
	FindByIDAndOwner(ctx context.Context, id ShaderID, owner UserID) (Shader, error)
	FindTop100ByCreatedByOrderByUpdatedAtDesc(ctx context.Context, createdBy UserID) ([]ShaderSummary, error)
	FindTop100PublicOrderByUpdatedAtDesc(ctx context.Context) ([]ShaderSummary, error)
	DeleteByIDAndOwner(ctx context.Context, id ShaderID, owner UserID) error
	DeleteByID(ctx context.Context, id ShaderID) error
	DeleteByOwner(ctx context.Context, owner UserID) error
	DeletePrivateByOwner(ctx context.Context, owner UserID) error
}
