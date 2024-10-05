package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
)

func ShaderCreate() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		var shaderCreate models.ShaderCreate
		err := json.NewDecoder(r.Body).Decode(&shaderCreate)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		shaderId, err := service.ShaderCreate(ctx, shaderCreate)
		if err != nil {
			return err
		}

		w.Header().Set("Location", fmt.Sprintf("/shader/%s", shaderId))
		w.WriteHeader(http.StatusCreated)

		return nil
	})
}

func ShaderById() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		shaderId := r.PathValue("id")
		if shaderId == "" {
			return infra.NotFoundError
		}

		switch r.Method {
		case "GET":
			shader, err := service.ShaderGet(ctx, shaderId)
			if err != nil {
				return err
			}

			shader.Location = fmt.Sprintf("/shader/%s", shader.Id)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(w).Encode(shader)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}
			return nil
		case "PUT":
			var shaderUpdate models.ShaderPartialUpdate
			err := json.NewDecoder(r.Body).Decode(&shaderUpdate)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}

			shader, err := service.ShaderUpdate(ctx, shaderId, shaderUpdate)
			if err != nil {
				return err
			}

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(shader)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}
			return nil
		default:
			return NewUnsupportedOperationError("GET", "PUT")
		}
	})
}

func ShaderInfoListOwn() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return NewUnsupportedOperationError("GET")
		}

		shaders, err := service.ShaderInfoListCurrentUser(ctx)
		if err != nil {
			return err
		}

		for idx, shader := range shaders {
			location := fmt.Sprintf("/shader/%s", shader.Id)
			shaders[idx].Location = location
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(w).Encode(shaders)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}
		return nil
	})
}
