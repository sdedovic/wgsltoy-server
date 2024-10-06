package db

import "github.com/sdedovic/wgsltoy-server/src/go/models"

type IRepository interface {
	UserCreate(username string, email string, hashedPassword string) (models.User, error)
	UserGetByUsername(username string) (models.User, error)
	UserGetById(userId string) (models.User, error)

	ShaderCreate(name string, visibility string, description string, tags []string, content string, createdBy string) (models.Shader, error)
	ShaderPartialUpdate(shaderId string, createdBy string, name *string, visibility *string, description *string, tags *[]string, content *string) (models.Shader, error)
	ShaderGetPubliclyVisibleById(shaderId string) (models.Shader, error)
	ShaderGetVisibleByIdAndLoggedInUser(shaderId string, currentUser string) (models.Shader, error)
	ShaderInfoListByCreatedBy(createdBy string) ([]models.ShaderInfo, error)
}
