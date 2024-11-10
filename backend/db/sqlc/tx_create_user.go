package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	// 在用户创建后在用一个事务中执行, 用它的返回值决定提交还是回滚
	AfterCreate func(user Users) error
}

type CreateUserTxResult struct {
	User Users
}

func (s *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	// 执行事务
	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
}
