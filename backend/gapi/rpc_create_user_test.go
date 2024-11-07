package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"simple_bank/pb"
	"simple_bank/pkg"
	"simple_bank/worker"
	mockwk "simple_bank/worker/mock"
	"testing"

	db "simple_bank/db/sqlc"

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
	arg      db.CreateUserTxParams
	password string

	user db.Users
}

func EqCreateUserParams(arg db.CreateUserTxParams, password string, user db.Users) gomock.Matcher {
	return eqMatcher{arg, password, user}
}

func (expected eqMatcher) Matches(x any) bool {
	// 将any 转换为 结构体
	// fmt.Println(">> check params")
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	// 校验hashed之后的密码
	// 比对两个密码的hash值, 如果相同则说明正确
	// fmt.Println(">> check password")
	err := pkg.CheckHashedPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	// 将预期参数的HashedPassword赋值给输入参数
	expected.arg.HashedPassword = actualArg.HashedPassword

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

	// 将预期参数与输入的参数对比, go无法直接比较函数, 使用参数里的结构体
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	// 用于模拟AfterCreate回调, 即用户成功创建之后执行该函数
	err = actualArg.AfterCreate(expected.user)
	return err == nil
}

func (expected eqMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", expected.arg, expected.password)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	require.NotEmpty(t, password)
	require.NotEmpty(t, user)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username:       user.Username,
						FullName:       user.FullName,
						HashedPassword: user.HashedPassword,
						Email:          user.Email,
					},
				}
				payload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserParams(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), payload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.User)
				require.NotNil(t, res.User.Username)
				require.NotNil(t, res.User.FullName)
			},
		}, {
			name: "Internal Error",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				// 判断是否来自内部错误
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			//  将store与taskDistributor的gomock控制器分开单独声明, 否则会造成锁问题
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			taskDistributorCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()

			store := mockdb.NewMockStore(storeCtrl)
			taskDistributor := mockwk.NewMockTaskDistributor(taskDistributorCtrl)

			tc.buildStubs(store, *taskDistributor)
			server := newTestServer(t, store, taskDistributor)

			// gRPC请求可以直接调用rpc函数并获取返回值
			response, err := server.CreateUser(context.Background(), tc.req)
			if err != nil {
				return
			}
			// 根据获取的返回值进入到校验响应函数
			tc.checkResponse(t, response, err)
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
