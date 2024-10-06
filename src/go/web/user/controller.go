package user

import (
	"context"
	"encoding/json"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/sdedovic/wgsltoy-server/src/go/service/user"
	"github.com/sdedovic/wgsltoy-server/src/go/web"
	"net/http"
)

type Controller struct {
	service user.IService `di.inject:"UserService"`
}

func (c *Controller) UserRegister() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return web.NewUnsupportedOperationError("POST")
		}

		// parse JSON
		var userRegister models.UserRegister
		err := json.NewDecoder(r.Body).Decode(&userRegister)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		err = c.service.Register(ctx, userRegister.Username, userRegister.Email, userRegister.Password)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

func (c *Controller) UserLogin() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return web.NewUnsupportedOperationError("POST")
		}

		var userLogin models.UserLogin
		err := json.NewDecoder(r.Body).Decode(&userLogin)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		jwt, err := c.service.Login(ctx, userLogin.Username, userLogin.Password)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(jwt))
		if err != nil {
			return err
		}

		return nil
	})
}

func (c *Controller) UserMe() http.HandlerFunc {
	return web.Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case "GET":
			currentUser, err := c.service.GetCurrent(ctx)
			if err != nil {
				return err
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(w).Encode(currentUser)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}
			return nil
		default:
			return web.NewUnsupportedOperationError("GET")
		}
	})
}
