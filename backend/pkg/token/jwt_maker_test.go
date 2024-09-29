package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"simple_bank/pkg"
)

var secretKey = pkg.RandomString(33)

func TestJWTMakerCreateToken(t *testing.T) {
	username := pkg.RandomString(5)
	duration := time.Minute * 5
	tokenString := randomCreateToken(t, username, duration)
	require.NotEmpty(t, tokenString)
}

func TestJWTMakerParseToken(t *testing.T) {
	maker := &JWTMaker{
		secretKey: []byte(secretKey),
	}
	t.Log(maker.secretKey)
	username := pkg.RandomString(5)
	duration := time.Minute * 5

	tokenString := randomCreateToken(t, username, duration)
	require.NotEmpty(t, tokenString)

	// 用例1: 使用正确的密钥测试校验函数
	payload, err := maker.VerifyToken(tokenString)
	t.Log(payload)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.Equal(t, []byte(secretKey), maker.secretKey)

	// 用例2: 使用错误的密钥测试校验函数
	maker2 := &JWTMaker{
		secretKey: []byte(pkg.RandomString(5)),
	}
	require.NotEqual(t, maker.secretKey, maker2.secretKey)
	tokenString2 := randomCreateToken(t, username, duration)
	require.NotEmpty(t, tokenString2)
	payload2, err2 := maker2.VerifyToken(tokenString2)
	require.Error(t, err2)
	require.Nil(t, payload2)
}

// 测试Token过期情况
func TestExpiredJWTToken(t *testing.T) {
	maker := &JWTMaker{
		secretKey: []byte(secretKey),
	}
	username := pkg.RandomString(5)
	duration := time.Minute * 6

	tokenString := randomCreateToken(t, username, -duration)
	require.NotEmpty(t, tokenString)

	// exp := time.Now().Add(time.Minute * 5)
	payload, err := maker.VerifyToken(tokenString)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

// 测试Alg None攻击
func TestInvalidJWTTokenAlgNone(t *testing.T) {
	tokenID := uuid.New()
	payload, err := NewPayload(tokenID, pkg.RandomString(5), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err2 := NewJWTMaker(secretKey)
	require.NoError(t, err2)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrTokenAlgNone.Error())
	require.Nil(t, payload)
}

func randomCreateToken(t *testing.T, username string, duration time.Duration) string {
	maker := &JWTMaker{
		secretKey: []byte(secretKey),
	}

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}
