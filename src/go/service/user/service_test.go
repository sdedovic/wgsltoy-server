package user

import (
	"context"
	"github.com/sdedovic/wgsltoy-server/src/go/db"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"github.com/sdedovic/wgsltoy-server/src/go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

type repoMock struct {
	db.IRepository

	userCreate        func(username string, email string, hashedPassword string) (models.User, error)
	userGetByUsername func(username string) (models.User, error)
	userGetById       func(userId string) (models.User, error)
}

func (m repoMock) UserCreate(username string, email string, hashedPassword string) (models.User, error) {
	return m.userCreate(username, email, hashedPassword)
}

func (m repoMock) UserGetByUsername(username string) (models.User, error) {
	return m.userGetByUsername(username)
}

func (m repoMock) UserGetById(userId string) (models.User, error) {
	return m.userGetById(userId)
}

func TestRegister_FailValidation(t *testing.T) {
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
			"one\u200Btwo\u200Bthree",
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
			mock := repoMock{
				userCreate: func(_, _, _ string) (models.User, error) {
					return models.User{}, nil
				},
			}
			s := &Service{mock}

			err := s.Register(context.Background(), tt.username, tt.email, tt.password)
			assert.IsType(t, infra.ValidationError{}, err)
		})
	}
}
