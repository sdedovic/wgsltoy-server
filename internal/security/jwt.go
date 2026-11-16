package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sdedovic/wgsltoy-server/internal/domain"
)

const issuer = "wgsltoy.com"
const tokenLifetime = 72 * time.Hour

var method = jwt.SigningMethodHS256

type JWT struct {
	secret []byte
}

var _ domain.TokenIssuer = (*JWT)(nil)
var _ domain.TokenVerifier = (*JWT)(nil)

func NewJWT(secret string) *JWT {
	return &JWT{[]byte(secret)}
}

func (j *JWT) Issue(identity domain.Identity) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": string(identity.UserID),
		"exp": now.Add(tokenLifetime).Unix(),
		"iat": now.Unix(),
		"iss": issuer,
	})
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed signing token caused by: %w", err)
	}

	return tokenString, nil
}

func (j *JWT) Verify(tokenString string) (domain.Identity, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{method.Name}))
	if err != nil || token == nil || !token.Valid {
		return domain.Identity{}, domain.ErrUnauthorized
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return domain.Identity{}, domain.ErrUnauthorized
	}

	userID, err := domain.NewUserID(subject)
	if err != nil {
		return domain.Identity{}, domain.ErrUnauthorized
	}

	return domain.Identity{UserID: userID}, nil
}
