package web

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/service"
	"net/http"
)

//==== User Register ====\\

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserRegister(pgPool *pgxpool.Pool) http.HandlerFunc {
	return CreateHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
