package shader

import (
	"context"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
)

type IService interface {
	ShaderCreate(ctx context.Context, shader models.ShaderCreate) (string, error)
	ShaderUpdate(ctx context.Context, shaderId string, shader models.ShaderPartialUpdate) (models.Shader, error)
	ShaderInfoListCurrentUser(ctx context.Context) ([]models.ShaderInfo, error)
	ShaderGet(ctx context.Context, shaderId string) (models.Shader, error)
}
