package api

import (
	"context"
	"simple_bank/pkg"
	"testing"

	"github.com/stretchr/testify/require"

	db "simple_bank/db/sqlc"
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

func randomTransferAccount(owner string) db.Accounts {
	// account, _ := testQueries.CreateAccount(context.Background(), db.CreateAccountParams{
	// 	Owner:    username,
	// 	Balance:  pkg.RandomInt(10, 100),
	// 	Currency: pkg.RandomCurrency(),
	// })
	// return account
	return db.Accounts{
		ID:       pkg.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  pkg.RandomInt(10, 100),
		Currency: pkg.RandomCurrency(),
	}
}

// TODO TestTransferAPI
// func TestTransferAPI(t *testing.T) {
// 	user1 := randomTransferUser(t)
// 	user2 := randomTransferUser(t)
// 	user3 := randomTransferUser(t)
//
// 	account1 := randomTransferAccount(user1.Username)
// 	account2 := randomTransferAccount(user2.Username)
// 	account3 := randomTransferAccount(user3.Username)
//
// 	amount := int64(10)
//
// 	account1.Currency = constants.CNY
// 	account2.Currency = constants.CNY
// 	account3.Currency = constants.CAD
//
// 	// transfer := db.CreateTransferParams{
// 	// 	FromAccountID: account1.ID,
// 	// 	ToAccountID:   account2.ID,
// 	// 	Amount:        10,
// 	// }
//
// 	testCases := []struct {
// 		name          string
// 		body          gin.H
// 		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
// 		buildStubs    func(store *mockdb.MockStore)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}{
// 		{
// 			name: "OK",
// 			body: gin.H{
// 				"fromAccountID": account1.ID,
// 				"toAccountID":   account2.ID,
// 				"amount":        amount,
// 				"currency":      constants.CNY,
// 			},
// 			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
// 				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, user1.Username, time.Minute)
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.
// 					EXPECT().
// 					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
// 					Times(1).
// 					Return(account1, nil)
//
// 				store.
// 					EXPECT().
// 					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
// 					Times(1).
// 					Return(account2, nil)
//
// 				arg := db.TransfersParams{
// 					FromAccountID: account1.ID,
// 					ToAccountID:   account2.ID,
// 					Amount:        amount,
// 				}
// 				store.
// 					EXPECT().
// 					TransferTx(gomock.Any(), gomock.Eq(arg)).
// 					Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusOK, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "两个账户的货币类型不匹配的情况",
// 			body: gin.H{
// 				"fromAccountID": account1.ID,
// 				"toAccountID":   account2.ID,
// 				"amount":        amount,
// 				"currency":      constants.CAD,
// 			},
// 			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
// 				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, user1.Username, time.Minute)
// 			},
// 			buildStubs: func(store *mockdb.MockStore) {
// 				store.
// 					EXPECT().
// 					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
// 					Times(1).
// 					Return(account1, nil)
//
// 				store.
// 					EXPECT().
// 					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
// 					Times(0)
// 				store.
// 					EXPECT().
// 					CreateTransfer(gomock.Any(), gomock.Any()).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				require.Equal(t, http.StatusForbidden, recorder.Code)
// 			},
// 		},
// 	}
//
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
//
// 			store := mockdb.NewMockStore(ctrl)
// 			tc.buildStubs(store)
//
// 			server := newTestServer(t, store)
// 			recorder := httptest.NewRecorder()
//
// 			body, err := json.Marshal(tc.body)
// 			require.NoError(t, err)
// 			data := bytes.NewReader(body)
//
// 			request := httptest.NewRequest(http.MethodPut, transRoute, data)
//
// 			server.router.ServeHTTP(recorder, request)
// 		})
// 	}
// }
