package web

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
	"time"
)

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserRegister(pgPool *pgxpool.Pool) http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		// parse JSON
		var registerUser RegisterUser
		err := json.NewDecoder(r.Body).Decode(&registerUser)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		err = service.UserRegister(ctx, pgPool, registerUser.Username, registerUser.Email, registerUser.Password)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func UserLogin(pool *pgxpool.Pool) http.HandlerFunc {
	return Handler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return NewUnsupportedOperationError("POST")
		}

		var loginUser LoginUser
		err := json.NewDecoder(r.Body).Decode(&loginUser)
		if err != nil {
			return infra.NewJsonParsingError(err)
		}

		jwt, err := service.UserLogin(ctx, pool, loginUser.Username, loginUser.Password)
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

type User struct {
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	EmailVerification string    `json:"emailVerificationStatus"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func UserMe(pool *pgxpool.Pool) http.HandlerFunc {
	return AuthenticatedHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		user := ctx.Value("user")
		if user == nil {
			return errors.New("no user in context")
		}
		userId := string(user.(service.UserInfo))

		switch r.Method {
		case "GET":

			user, err := service.UserFindOne(ctx, pool, userId)
			if err != nil {
				return err
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(w).Encode(User{user.Username, user.Email, user.EmailVerification, user.CreatedAt, user.UpdatedAt})
			if err != nil {
				return infra.NewJsonParsingError(err)
			}
			return nil

		default:
			return NewUnsupportedOperationError("GET")
		}
	})
}
