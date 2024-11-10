package db

import "context"

type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User        Users
	VerifyEmail VerifyEmails
}

func (s *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	// 执行事务
	// 1. 通过emailId和secret_code找到验证邮件
	// 2. 将is_used更新为true
	// 3. 将is_email_verified更新为true
	err := s.execTx(ctx, func(q *Queries) error {
		var err error
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		type Ok struct {
			Ok bool
		}
		ok := Ok{Ok: true}

		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username:        &result.VerifyEmail.Username,
			IsEmailVerified: &ok.Ok,
		})
		if err != nil {
			return err
		}
		return err
	})

	return result, err
}
