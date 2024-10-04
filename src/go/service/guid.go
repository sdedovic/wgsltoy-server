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

func ValidateGUID(guid string) bool {
	if len(guid) != 22 {
		return false
	}

	bytes, err := base64.RawURLEncoding.Strict().DecodeString(guid)
	if err != nil {
		return false
	}

	v4, err := uuid.FromBytes(bytes)
	if err != nil {
		return false
	}

	return v4.Version() == 4
}
