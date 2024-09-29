package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	mockdb "simple_bank/db/mock"

	db "simple_bank/db/sqlc"
	"simple_bank/pkg"

	"go.uber.org/mock/gomock"
)

func TestAccountAPI(t *testing.T) {
	// mock random account
	account := randomAccount()
	fmt.Printf("account %v", account)

	// 测试用例集合, 存储数据可能的情况
	// name: 测试用例名
	// accountID: 账户ID
	// buildStubs: 测试函数
	// checkResponse: 检查HTTP Response的函数
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
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
		},
		{
			name:      "NotFoundID",
			accountID: account.ID,
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
		{
			name: "BadRequest",
			// accountID: account.ID,
			accountID: -1,
			buildStubs: func(store *mockdb.MockStore) {
				// 因为ID无效, gin校验时直接返回错误, 并没有调用GetAccount, 也就没有返回值
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Accounts {
	return db.Accounts{
		ID:       pkg.RandomInt(1, 10),
		Owner:    pkg.RandomString(5),
		Balance:  pkg.RandomInt(1, 100),
		Currency: pkg.RandomCurrency(),
	}
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
