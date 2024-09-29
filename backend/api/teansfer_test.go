package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"simple_bank/constants"
	"testing"

	"simple_bank/pkg"

	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"

	db "simple_bank/db/sqlc"

	"go.uber.org/mock/gomock"

	mockdb "simple_bank/db/mock"
)

const transRoute = "/transfers"

func randomTransferUser(t *testing.T) db.Users {
	hashedPassword, err := pkg.HashPassword(pkg.RandomString(6))
	require.NoError(t, err)
	require.NotNil(t, hashedPassword)

	user, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
		Username:       pkg.RandomString(5),
		FullName:       pkg.RandomString(10),
		HashedPassword: hashedPassword,
		Email:          pkg.RandomEmail(3),
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotEmpty(t, user)

	return user
}

func randomTransferAccount(t *testing.T) db.Accounts {
	user := randomTransferUser(t)
	account, err := testQueries.CreateAccount(context.Background(), db.CreateAccountParams{
		Owner:    user.Username,
		Balance:  pkg.RandomInt(10, 100),
		Currency: pkg.RandomCurrency(),
	})
	require.NoError(t, err)
	require.NotNil(t, account)
	require.NotEmpty(t, account)

	return account
}

func TestTransferAPI(t *testing.T) {
	account1 := randomTransferAccount(t)
	account2 := randomTransferAccount(t)
	amount := int64(10)
	account1.Currency = constants.CNY
	account2.Currency = constants.CNY

	// transfer := db.CreateTransferParams{
	// 	FromAccountID: account1.ID,
	// 	ToAccountID:   account2.ID,
	// 	Amount:        10,
	// }

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"fromAccountID": account1.ID,
				"toAccountID":   account2.ID,
				"amount":        amount,
				"currency":      constants.CNY,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				arg := db.CreateTransferParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}
				store.
					EXPECT().
					CreateTransfer(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "两个账户的货币类型不匹配的情况",
			body: gin.H{
				"fromAccountID": account1.ID,
				"toAccountID":   account2.ID,
				"amount":        amount,
				"currency":      constants.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.
					EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(0)
				store.
					EXPECT().
					CreateTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			data := bytes.NewReader(body)

			request := httptest.NewRequest(http.MethodPut, transRoute, data)

			server.router.ServeHTTP(recorder, request)
		})
	}
}
