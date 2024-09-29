package token

import (
	"time"
)

// Maker 通用的Token接口管理令牌的颁发和校验, 用于切换不同的Token类型
type Maker interface {
	// CreateToken 用户名与过期时间, 对特定用户的令牌或有效时期进行颁发
	CreateToken(username string, duration time.Duration) (string, error)
	// VerifyToken 验证token是否合法
	VerifyToken(token string) (*Payload, error)
}
