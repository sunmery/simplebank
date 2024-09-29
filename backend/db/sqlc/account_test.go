package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"simple_bank/pkg"
)

func TestCreateAccount(t *testing.T) {
	account := createRandomAccount(t)
	require.NotNil(t, account)
}

func TestGetAccount(t *testing.T) {
	ctx := context.Background()
	account1 := createRandomAccount(t)
	account2, err := testQueries.GetAccount(ctx, account1.ID)
	require.NoError(t, err)
	require.NotNil(t, account2)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
}

func TestUpdateAccount(t *testing.T) {
	ctx := context.Background()
	account := createRandomAccount(t)

	balance := pkg.RandomInt(1, 100)
	result, err := testQueries.UpdateAccount(ctx, UpdateAccountParams{
		Balance: balance,
		ID:      account.ID,
	})
	require.NotNil(t, result)
	require.NoError(t, err)

	require.Equal(t, balance, result.Balance)

	require.NotZero(t, result.ID)
	require.NotZero(t, result.CreatedAt)
}

func TestListAccount(t *testing.T) {
	var lastAccount Accounts
	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}
	accounts, err := testQueries.ListAccounts(context.Background(), ListAccountsParams{
		// Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.NotEmpty(t, lastAccount)
	}
}

func createRandomAccount(t *testing.T) Accounts {
	ctx := context.Background()
	user := createRandomUser(t)
	require.NotEmpty(t, user)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  pkg.RandomInt(11, 100),
		Currency: pkg.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(ctx, arg)
	if err != nil {
		t.Fatalf("account generate error is: '%v'", err)
	}
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}
