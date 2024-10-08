package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"simple_bank/constants"
	"simple_bank/pkg/token"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	mockdb "simple_bank/db/mock"

	db "simple_bank/db/sqlc"
	"simple_bank/pkg"

	"go.uber.org/mock/gomock"
)

func TestAccountAPI(t *testing.T) {
	username := pkg.RandomString(5)
	// mock random account
	account := randomAccount(t, username)
	fmt.Printf("account %v", account)

	// 测试用例集合, 存储数据可能的情况
	// name: 测试用例名
	// accountID: 账户ID
	// buildStubs: 测试函数
	// checkResponse: 检查HTTP Response的函数
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		}, {
			name:      "令牌的username与用户的username不匹配",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, "Unauthorized username", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		}, {
			name:      "未授权",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFoundID",
			accountID: int64(-1),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, username, time.Minute)
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Accounts{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	// start test web http server & send request
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// test api
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMake)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(t *testing.T, username string) db.Accounts {
	account := db.Accounts{
		ID:       pkg.RandomInt(1, 10),
		Owner:    username,
		Balance:  pkg.RandomInt(1, 100),
		Currency: pkg.RandomCurrency(),
	}
	require.NotEmpty(t, account)
	require.NotZero(t, account.ID)
	require.NotEmpty(t, account.Owner)
	require.NotZero(t, account.Balance)
	require.NotEmpty(t, account.Currency)

	return account
}

// 测试Body体的数据是否与生成的账户数据相同
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Accounts) {
	// http req body的类型是 bytes.Buffer 需要解构为结构体然后比对
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var dotAccount db.Accounts
	err = json.Unmarshal(data, &dotAccount)
	require.NoError(t, err)

	require.Equal(t, dotAccount, account)
}
