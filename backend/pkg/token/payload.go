package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
)

type Payload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`

	jwt.RegisteredClaims
}

// NewPayload 创建一个荷载
// 创建一个uuid, 用于管理token
func NewPayload(id uuid.UUID, username string, duration time.Duration) (*Payload, error) {
	// 生成随机的uuid
	payload := &Payload{
		ID:       id,
		Username: username,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        id.String(),
		},
	}

	return payload, nil
}

// Valid 校验Token是否过期
func (payload *Payload) Valid() error {
	if time.Now().After(payload.RegisteredClaims.ExpiresAt.Time) {
		return ErrExpiredToken
	}
	return nil
}
