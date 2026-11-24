//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package server

type Controller interface {
	Register(router *Router)
}
