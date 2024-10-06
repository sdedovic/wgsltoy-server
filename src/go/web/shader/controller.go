package shader

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/sdedovic/wgsltoy-server/src/go/service/shader"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"net/http"
)

type Controller struct {
	service shader.IService `di.inject:"ShaderService"`
}

func (c *Controller) ShaderCreate() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return web.NewUnsupportedOperationError("POST")
		}

		var shaderCreate models.ShaderCreate
		err := json.NewDecoder(r.Body).Decode(&shaderCreate)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		shaderId, err := c.service.ShaderCreate(ctx, shaderCreate)
		if err != nil {
			return err
		}

		w.Header().Set("Location", fmt.Sprintf("/shader/%s", shaderId))
		w.WriteHeader(http.StatusCreated)

		return nil
	})
}

func (c *Controller) ShaderById() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		shaderId := r.PathValue("id")
		if shaderId == "" {
			return infra.NotFoundError
		}

		switch r.Method {
		case "GET":
			shader, err := c.service.ShaderGet(ctx, shaderId)
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

			shader, err := c.service.ShaderUpdate(ctx, shaderId, shaderUpdate)
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
			return web.NewUnsupportedOperationError("GET", "PUT")
		}
	})
}

func (c *Controller) ShaderInfoListOwn() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return web.NewUnsupportedOperationError("GET")
		}

		shaders, err := c.service.ShaderInfoListCurrentUser(ctx)
		if err != nil {
			return err
		}

		for idx, s := range shaders {
			location := fmt.Sprintf("/shader/%s", s.Id)
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
