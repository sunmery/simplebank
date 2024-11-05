package gapi

import (
	"context"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	db "simple_bank/db/sqlc"
	"simple_bank/pb"
	"simple_bank/validator"
)

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	// 校验rpc参数
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidCounterargument(violations)
	}

	verifyEmailTx, err := s.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		// TODO: 根据不同状态码来返回不同的错误
		return nil, status.Errorf(codes.Internal, "无法验证电子邮件: %v", err)
	}

	return &pb.VerifyEmailResponse{
		IsVerified: verifyEmailTx.User.IsEmailVerified,
	}, nil
}

// 校验验证邮件RPC参数
func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}
	// 应为proto中定义的字段名, 即蛇形命名法的字段
	if err := validator.ValidateEmailSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}
	return
}
