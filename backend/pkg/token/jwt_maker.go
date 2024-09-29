package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecretKeyMinLength = 32

var (
	ErrExpiredToken = errors.New("token has invalid claims: token is expired")
	ErrTokenAlgNone = errors.New("token signature is invalid: token is unverifiable: 'none' signature type is not allowed")
)

type JWTMaker struct {
	secretKey []byte
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < jwtSecretKeyMinLength {
		return nil, errors.New(fmt.Sprintf("jwt secret key长度不能少于 %d", jwtSecretKeyMinLength))
	}
	return &JWTMaker{secretKey: []byte(secretKey)}, nil
}

// CreateToken 用户名与过期时间, 对特定用户的令牌或有效时期进行颁发
func (maker JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	tokenID := uuid.New()
	claims, err := NewPayload(tokenID, username, duration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(maker.secretKey)
}

// VerifyToken 验证token是否合法
func (maker JWTMaker) VerifyToken(tokenString string) (*Payload, error) {
	funcKey := func(t *jwt.Token) (interface{}, error) {
		return maker.secretKey, nil
	}
	token, err := jwt.ParseWithClaims(tokenString, &Payload{}, funcKey)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Payload); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}
