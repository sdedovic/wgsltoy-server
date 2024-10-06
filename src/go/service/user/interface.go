package user

import (
	"context"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
)

type IService interface {
	Register(ctx context.Context, username string, email string, password string) error
	Login(ctx context.Context, username string, password string) (string, error)
	GetCurrent(ctx context.Context) (models.User, error)
}
