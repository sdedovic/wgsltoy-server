package web

import (
	"context"
	"encoding/json"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"net/http"
)

//==== Health Check ====\\

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthCheck() http.HandlerFunc {
	return CreateHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return NewUnsupportedOperationError("GET")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(HealthResponse{"ok"})
		if err != nil {
			return infra.NewJsonParsingError(err)
		}
		return nil
	})
}
