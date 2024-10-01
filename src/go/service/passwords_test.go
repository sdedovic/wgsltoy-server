package service

import "testing"

func TestHashPassword(t *testing.T) {
	password := "$foo-bar123!"

	// create password hash
	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("error hashing password caused by: %v", err)
	}

	// true when checking correct password
	isMatch, err := VerifyPassword(password, passwordHash)
	if err != nil {
		t.Fatalf("error verifying password caused by: %v", err)
	}
	if !isMatch {
		t.Fatalf("password does not match")
	}

	// false when checking wrong password
	isMatch, err = VerifyPassword("this should fail", passwordHash)
	if err != nil {
		t.Fatalf("error verifying password caused by: %v", err)
	}
	if isMatch {
		t.Fatalf("password must not match")
	}
}
