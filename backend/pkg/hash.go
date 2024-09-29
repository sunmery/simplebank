package pkg

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 根据迭代次数参数与盐值生成的密码散列字符串
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("faild genate hashed password err is '%w'", err)
	}
	return string(hashedPassword), nil
}

// CheckHashedPassword 校验散列密码字符串
func CheckHashedPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
