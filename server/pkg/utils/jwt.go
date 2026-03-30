package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5" // 导入 v5
	"github.com/google/uuid"
)

// TokenType 标识 Token 类型
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// MyClaims 自定义 JWT Claims
type MyClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType string    `json:"token_type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

// CreateAccessToken 创建短期 Access Token
func CreateAccessToken(userID uuid.UUID, secret []byte, expire time.Duration) (string, error) {
	return createToken(userID, secret, expire, TokenTypeAccess)
}

// CreateRefreshToken 创建长期 Refresh Token
func CreateRefreshToken(userID uuid.UUID, secret []byte, expire time.Duration) (string, error) {
	return createToken(userID, secret, expire, TokenTypeRefresh)
}

// CreateTokenPair 一次生成双 Token
func CreateTokenPair(userID uuid.UUID, secret []byte, accessExpire, refreshExpire time.Duration) (accessToken, refreshToken string, err error) {
	accessToken, err = CreateAccessToken(userID, secret, accessExpire)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = CreateRefreshToken(userID, secret, refreshExpire)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ParseToken 解析并验证 Token
func ParseToken(tokenStr string, secret []byte) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// 类型断言并验证
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// createToken 内部方法，创建指定类型的 Token
func createToken(userID uuid.UUID, secret []byte, expire time.Duration, tokenType string) (string, error) {
	claims := MyClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
