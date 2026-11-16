//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package server

import "net/http"

type Controller interface {
	Register(mux *http.ServeMux)
}
