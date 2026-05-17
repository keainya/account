package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTSecret      = "account-system-jwt-secret-change-in-production"
	AccessTokenTTL  = time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour
)

// Claims JWT 声明
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

// 用于 metadata 中间件的 context key
type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUsername contextKey = "username"
	ContextKeyRole     contextKey = "role"
	ContextKeyClientID contextKey = "client_id"
	ContextKeyUser     contextKey = "user"
)

// GenerateAccessToken 生成 access_token
func GenerateAccessToken(userID, username, role, clientID string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
