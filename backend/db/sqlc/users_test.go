package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"simple_bank/pkg"
)

var testQueries *Queries

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, user1.Username, user.Username)
}

func createRandomUser(t *testing.T) Users {
	ctx := context.Background()

	username := pkg.RandomString(5)
	fullName := pkg.RandomString(5)
	hashedPassword := pkg.RandomString(5)
	email := pkg.RandomEmail(5)
	arg := CreateUserParams{
		Username:       username,
		FullName:       fullName,
		HashedPassword: hashedPassword,
		Email:          email,
	}

	user, err := testQueries.CreateUser(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, user.Username, arg.Username)
	require.Equal(t, user.FullName, arg.FullName)
	require.Equal(t, user.Email, arg.Email)
	require.NotZero(t, user.PasswordChangedAt)
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.UpdatedAt)

	return user
}
