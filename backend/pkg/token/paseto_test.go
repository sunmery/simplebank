package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"simple_bank/pkg"
)

var symmetricKey = pkg.RandomString(32)

func TestCreateToken(t *testing.T) {
	token := createToken(t)
	require.NotEmpty(t, token)
	require.NotNil(t, token)
}

func TestVerifyToken(t *testing.T) {
	token := createToken(t)
	require.NotEmpty(t, token)
	require.NotNil(t, token)

	maker, err := NewPasetoMaker(symmetricKey)
	require.NoError(t, err)
	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotNil(t, payload)
	t.Log(payload)
}

func createToken(t *testing.T) string {
	maker, err := NewPasetoMaker(symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)
	require.NotEmpty(t, maker)

	username := pkg.RandomString(5)
	token, err := maker.CreateToken(username, time.Minute*2)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	return token
}
