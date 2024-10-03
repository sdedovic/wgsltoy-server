package service

import (
	"encoding/base64"
	"github.com/google/uuid"
)

// NewGUID generates a 22 character long globally unique identifier
func NewGUID() string {
	v4, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.Strict().EncodeToString(v4[:])
}
