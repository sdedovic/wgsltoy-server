package web

import (
	"context"
	"encoding/json"
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
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

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
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "PUT" {
			return NewUnsupportedOperationError("PUT")
		}

		shaderId := r.PathValue("id")
		if shaderId == "" {
			return infra.NotFoundError
		}

		var updateShader *UpdateShader
		err := json.NewDecoder(r.Body).Decode(&updateShader)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		err = service.ShaderUpdate(ctx, pgPool, shaderId, service.UpdateShader{
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

type ShaderInfo struct {
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
}

func ShaderInfoListOwn(pgPool *pgxpool.Pool) http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return NewUnsupportedOperationError("GET")
		}

		shaders, err := service.ShaderInfoListCurrentUser(ctx, pgPool)
		if err != nil {
			return err
		}

		output := make([]ShaderInfo, len(shaders))
		for i, shader := range shaders {
			location := fmt.Sprintf("/shader/%s", shader.Id)

			forkedFromLocation := ""
			if shader.ForkedFrom != "" {
				forkedFromLocation = fmt.Sprintf("/shader/%s", shader.ForkedFrom)
			}

			output[i] = ShaderInfo{
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
