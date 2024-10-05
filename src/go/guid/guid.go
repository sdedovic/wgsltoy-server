package guid

import (
	"encoding/base64"
	"github.com/google/uuid"
)

const RuneLength = 22

// New creates a new unique identifier or panics
func New() string {
	v4, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.Strict().EncodeToString(v4[:])
}

// Validate returns true if the supplied value is a valid GUID
func Validate(s string) bool {
	if len(s) != 22 {
		return false
	}

	bytes, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil {
		return false
	}

	v4, err := uuid.FromBytes(bytes)
	if err != nil {
		return false
	}

	return v4.Version() == 4
}
