package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"simple_bank/pkg"

	db "simple_bank/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	mockdb "simple_bank/db/mock"

	"go.uber.org/mock/gomock"
)

type Matcher interface {
	// Matches returns whether x is a match.
	Matches(x any) bool

	// String describes what the matcher matches.
	String() string
}

type eqMatcher struct {
	arg      db.CreateUserParams
	password string
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqMatcher{arg, password}
}

func (e eqMatcher) Matches(x any) bool {
	// 将any 转换为 结构体
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	// 校验hashed之后的密码
	// 比对两个密码的hash值, 如果相同则说明正确
	err := pkg.CheckHashedPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	// 将预期参数的HashedPassword赋值给输入参数
	e.arg.HashedPassword = arg.HashedPassword

	// // In case, some value is nil
	// if e.arg == db.CreateUserParams || x == nil {
	// 	return reflect.DeepEqual(e.x, x)
	// }
	//
	// // Check if types assignable and convert them to common type
	// x1Val := reflect.ValueOf(e.x)
	// x2Val := reflect.ValueOf(x)
	//
	// if x1Val.Type().AssignableTo(x2Val.Type()) {
	// 	x1ValConverted := x1Val.Convert(x2Val.Type())
	// 	return reflect.DeepEqual(x1ValConverted.Interface(), x2Val.Interface())
	// }
	//
	// return false

	// 将预期参数与输入的参数对比
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", e.arg, e.password)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	require.NotEmpty(t, password)
	require.NotEmpty(t, user)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"fullName": user.FullName,
				"password": password,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					FullName:       user.FullName,
					HashedPassword: user.HashedPassword,
					Email:          user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(db.Users{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Password Length < 6",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike@example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Bad email",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Create duplicate users",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike@example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			// 解构body io流的body data为JSON
			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			data := bytes.NewReader(body)

			// test uses api
			url := "/users"
			request, readErr := http.NewRequest(http.MethodPut, url, data)
			require.NoError(t, readErr)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder)
		})
	}
}

// TODO TestGetUserAPI
func TestGetUserAPI(t *testing.T) {
	user, password := randomUser(t)
	require.NotEmpty(t, password)
	require.NotEmpty(t, user)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"fullName": user.FullName,
				"password": password,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					FullName:       user.FullName,
					HashedPassword: user.HashedPassword,
					Email:          user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(db.Users{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Password Length < 6",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike@example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Bad email",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Create duplicate users",
			body: gin.H{
				"username": "mike",
				"password": "mike1",
				"fullName": "Mike",
				"email":    "mike@example.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			// 解构body io流的body data为JSON
			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			data := bytes.NewReader(body)

			// test uses api
			url := "/users"
			request, readErr := http.NewRequest(http.MethodPut, url, data)
			require.NoError(t, readErr)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.Users, password string) {
	password = pkg.RandomString(6)

	hashedPassword, err := pkg.HashPassword(password)
	require.NoError(t, err)

	user = db.Users{
		Username:       pkg.RandomString(2),
		FullName:       pkg.RandomString(12),
		HashedPassword: hashedPassword,
		Email:          pkg.RandomEmail(5),
	}
	return
}
