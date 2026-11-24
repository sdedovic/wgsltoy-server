package security

import (
	"testing"

	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

func TestArgon2Hasher(t *testing.T) {
	hasher := NewArgon2Hasher()
	password := domain.Password("$foo-bar123!")

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("error hashing password caused by: %v", err)
	}

	isMatch, err := hasher.Verify(password, passwordHash)
	if err != nil {
		t.Fatalf("error verifying password caused by: %v", err)
	}
	if !isMatch {
		t.Fatalf("password does not match")
	}

	isMatch, err = hasher.Verify("this should fail", passwordHash)
	if err != nil {
		t.Fatalf("error verifying password caused by: %v", err)
	}
	if isMatch {
		t.Fatalf("password must not match")
	}
}
