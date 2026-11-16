package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockUserRepository(ctrl)
	hasher := NewMockPasswordHasher(ctrl)

	hasher.EXPECT().Hash(Password("valid-password123")).Return("hashed", nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, user User) error {
		assert.Equal(t, Username("TestUser3"), user.Username)
		assert.Equal(t, Email("testuser@wgsltoy.com"), user.Email)
		assert.Equal(t, EmailVerificationPending, user.EmailVerification)
		assert.Equal(t, "hashed", user.PasswordHash)
		assert.NotEmpty(t, user.ID)
		return nil
	})

	s := NewUserService(repo, hasher, NewMockTokenIssuer(ctrl))
	err := s.Register(context.Background(), "TestUser3", "testuser@wgsltoy.com", "valid-password123")
	assert.NoError(t, err)
}

func TestUserService_Register_FailValidation(t *testing.T) {
	tests := []struct {
		name string

		username string
		email    string
		password string
	}{
		{"username required",
			"",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username banned",
			"admin",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username too short",
			"x",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username too long",
			"x111111122222223333333333333333",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username invalid characters",
			"x o 1",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username invalid characters",
			"; DROP TABLE users;",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"username invalid characters",
			"one​two​three",
			"admin@wgsltoy.com",
			"valid-password123",
		},
		{"email required",
			"TestUser3",
			"",
			"valid-password123",
		},
		{"email invalid",
			"TestUser3",
			"foo",
			"valid-password123",
		},
		{
			"password required",
			"TestUser9",
			"testuser@wgsltoy.com",
			"",
		},
		{
			"password too short",
			"TestUser9",
			"testuser@wgsltoy.com",
			"abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no expectations, so any dependency call fails the test
			ctrl := gomock.NewController(t)
			s := NewUserService(NewMockUserRepository(ctrl), NewMockPasswordHasher(ctrl), NewMockTokenIssuer(ctrl))

			err := s.Register(context.Background(), tt.username, tt.email, tt.password)
			assert.IsType(t, ValidationError{}, err)
		})
	}
}
