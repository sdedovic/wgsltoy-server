package web

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
)

type CreateShader struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
	ForkedFrom  string   `json:"forkedFrom"`
}

func Shader(pgPool *pgxpool.Pool) http.HandlerFunc {
	return AuthenticatedHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		user := ctx.Value("user")
		if user == nil {
			return errors.New("no user in context")
		}
		userId := string(user.(service.UserInfo))

		switch r.Method {
		case "POST":
			var createShader *CreateShader
			err := json.NewDecoder(r.Body).Decode(&createShader)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}

			_, err = service.ShaderCreate(ctx, pgPool, &service.CreateShader{
				Name:        createShader.Name,
				Visibility:  createShader.Visibility,
				Description: createShader.Description,
				Content:     createShader.Content,
				Tags:        createShader.Tags,
				CreatedBy:   userId,
				ForkedFrom:  createShader.ForkedFrom,
			})
			if err != nil {
				return err
			}

			return nil
		default:
			return NewUnsupportedOperationError("POST")
		}
	})
}
