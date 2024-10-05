package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"os"
	"time"
)

const issuer = "wgsltoy.com"
const ContextKey = "user"

var method = jwt.SigningMethodHS256

type UserInfo struct {
	Id string
}

func ExtractUserInfoFromContext(ctx context.Context) *UserInfo {
	value := ctx.Value(ContextKey)
	switch v := value.(type) {
	case *UserInfo:
		return v
	default:
		return nil
	}
}

func InsertUserInfoIntoContext(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, ContextKey, userInfo)
}

func MakeToken(user UserInfo) (string, error) {
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
		"iss": issuer,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("APP_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed signing token caused by: %w", err)
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("APP_SECRET")), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{method.Name}))

	if err != nil || token == nil || !token.Valid {
		return nil, infra.UnauthorizedError
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("failed extracting subject from token caused by: %w", err)
	}

	return &UserInfo{subject}, nil
}
