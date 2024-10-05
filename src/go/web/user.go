package web

import (
	"context"
	"encoding/json"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
)

func UserRegister() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		// parse JSON
		var userRegister models.UserRegister
		err := json.NewDecoder(r.Body).Decode(&userRegister)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		err = service.UserRegister(ctx, userRegister.Username, userRegister.Email, userRegister.Password)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

func UserLogin() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		var userLogin models.UserLogin
		err := json.NewDecoder(r.Body).Decode(&userLogin)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		jwt, err := service.UserLoginGenerateToken(ctx, userLogin.Username, userLogin.Password)
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

func UserMe() http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case "GET":
			user, err := service.UserGetCurrent(ctx)
			if err != nil {
				return err
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(w).Encode(user)
			if err != nil {
				return infra.NewJsonParsingError(err)
			}
			return nil
		default:
			return NewUnsupportedOperationError("GET")
		}
	})
}
