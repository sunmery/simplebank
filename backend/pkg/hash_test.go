package pkg

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := RandomString(10)
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// 正确用例: 正确地校验, 说明生成的散列密码与密码对应
	err = CheckHashedPassword(password, hashedPassword)
	require.NoError(t, err)

	// 错误用例: 当发生错误时, 应与CheckHashedPassword校验函数的错误一致:
	wrongPassword := RandomString(10)
	err = CheckHashedPassword(wrongPassword, hashedPassword)
	// bcrypt.ErrMismatchedHashAndPassword.Error() 是校验错误的错误函数
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	// 用例3: 保证相同的密码产生不同的散列字符串
	password2, err2 := HashPassword(password)
	require.NoError(t, err2)
	require.NotEmpty(t, password2)
	require.NotEqual(t, password, password2)
}
