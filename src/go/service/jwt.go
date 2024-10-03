package service

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sdedovic/wgsltoy-server/src/go/infra"
	"os"
	"time"
)

const issuer = "wgsltoy.com"

var method = jwt.SigningMethodHS256

type UserInfo string

func MakeToken(user UserInfo) (string, error) {
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": user,
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

func ParseToken(tokenString string) (UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("APP_SECRET")), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{method.Name}))

	if err != nil || token == nil || !token.Valid {
		return "", infra.UnauthorizedError
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("failed extracting subject from token: %w", err)
	}

	return UserInfo(subject), nil
}
