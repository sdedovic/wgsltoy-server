package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
	"time"
)

type CreateShader struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	ForkedFrom  string   `json:"forkedFrom"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
}

func ShaderCreate(pgPool *pgxpool.Pool) http.HandlerFunc {
	return AuthenticatedHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		user := ctx.Value("user")
		if user == nil {
			return errors.New("no user in context")
		}
		userId := string(user.(service.UserInfo))

		var createShader *CreateShader
		err := json.NewDecoder(r.Body).Decode(&createShader)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		shaderId, err := service.ShaderCreate(ctx, pgPool, &service.CreateShader{
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

		w.Header().Set("Location", fmt.Sprintf("/shader/%s", shaderId))
		w.WriteHeader(http.StatusCreated)

		return nil
	})
}

type UpdateShader struct {
	Name        string   `json:"name"`
	Visibility  string   `json:"visibility"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
}

func ShaderUpdate(pgPool *pgxpool.Pool) http.HandlerFunc {
	return AuthenticatedHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "PUT" {
			return NewUnsupportedOperationError("PUT")
		}

		user := ctx.Value("user")
		if user == nil {
			return errors.New("no user in context")
		}
		userId := string(user.(service.UserInfo))

		shaderId := r.PathValue("id")
		if shaderId == "" {
			return infra.NotFoundError
		}

		var updateShader *UpdateShader
		err := json.NewDecoder(r.Body).Decode(&updateShader)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		err = service.ShaderUpdate(ctx, pgPool, userId, shaderId, service.UpdateShader{
			Name:        updateShader.Name,
			Visibility:  updateShader.Visibility,
			Description: updateShader.Description,
			Tags:        updateShader.Tags,
			Content:     updateShader.Content,
		})
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

type Shader struct {
	Id        string    `json:"id"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name               string   `json:"name"`
	Visibility         string   `json:"visibility"`
	Description        string   `json:"description"`
	ForkedFrom         string   `json:"forkedFrom,omitempty"`
	ForkedFromLocation string   `json:"forkedFromLocation,omitempty"`
	Tags               []string `json:"tags"`
	Content            string   `json:"content"`
}

func ShaderListOwn(pgPool *pgxpool.Pool) http.HandlerFunc {
	return AuthenticatedHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return NewUnsupportedOperationError("GET")
		}

		user := ctx.Value("user")
		if user == nil {
			return errors.New("no user in context")
		}
		userId := string(user.(service.UserInfo))

		shaders, err := service.ShaderListByUser(ctx, pgPool, userId)
		if err != nil {
			return err
		}

		output := make([]Shader, len(shaders))
		for i, shader := range shaders {
			location := fmt.Sprintf("/shader/%s", shader.Id)

			forkedFromLocation := ""
			if shader.ForkedFrom != "" {
				forkedFromLocation = fmt.Sprintf("/shader/%s", shader.ForkedFrom)
			}

			output[i] = Shader{
				shader.Name,
				location,
				shader.CreatedAt,
				shader.UpdatedAt,
				shader.Name,
				shader.Visibility,
				shader.Description,
				shader.ForkedFrom,
				forkedFromLocation,
				shader.Tags,
				shader.Content,
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(w).Encode(output)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}
		return nil
	})
}
