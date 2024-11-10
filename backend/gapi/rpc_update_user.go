package gapi

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	db "simple_bank/db/sqlc"
	"simple_bank/pb"
	"simple_bank/pkg"
	"simple_bank/validator"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 校验rpc参数
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidCounterargument(violations)
	}

	var password string
	if req.Password != nil {
		var err error
		password, err = pkg.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "无法获取密码")
		}
	}

	arg := db.UpdateUserParams{
		Username:       &req.Username,
		FullName:       req.FullName,
		HashedPassword: &password,
		Email:          req.Email,
	}
	user, err := s.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "未查找到用户名 %s", err)
		}
		return nil, status.Errorf(codes.Internal, "未知错误 %s", err)
	}

	return &pb.UpdateUserResponse{
		User: &pb.User{
			Username:         user.Username,
			FullName:         user.FullName,
			Email:            user.Email,
			PasswordChangeAt: timestamppb.New(user.PasswordChangedAt),
		},
	}, nil
}

// 校验注册用户RPC参数
func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if req.FullName != nil {

		// 应为proto中定义的字段名, 即蛇形命名法的字段
		if err := validator.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}
	if req.Password != nil {
		if err := validator.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	if req.Email != nil {
		if err := validator.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	return
}
