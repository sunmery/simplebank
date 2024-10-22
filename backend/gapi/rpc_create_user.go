package gapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	db "simple_bank/db/sqlc"
	"simple_bank/pb"
	"simple_bank/pkg"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	password, err := pkg.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "无法获取密码")
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		FullName:       req.GetFullName(),
		HashedPassword: password,
		Email:          req.GetEmail(),
	}
	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, status.Errorf(codes.AlreadyExists, "用户名已存在 %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "未知错误 %s", err)
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Username:         user.Username,
			FullName:         user.FullName,
			Email:            user.Email,
			PasswordChangeAt: timestamppb.New(user.PasswordChangedAt),
			CreateAt:         timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	// 查询客户端传递的username参数
	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		// 查不到
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "该用户不存在: %s", err)
		}
		// 其它错误
		return nil, status.Errorf(codes.Internal, "未知错误: %s", err)
	}

	// 检查密码与hash之后的密码是否匹配
	if checkErr := pkg.CheckHashedPassword(req.Password, user.HashedPassword); checkErr != nil {
		// 401
		return nil, status.Errorf(codes.NotFound, "密码不匹配, 未通过身份验证: %s", checkErr)
	}

	// 颁发token
	token, accessPayload, createErr := s.tokenMake.CreateToken(user.Username, s.config.AccessTokenDuration)
	if createErr != nil {
		return nil, status.Errorf(codes.Internal, "颁发token失败, 内部错误: %s", createErr)
	}

	// 刷新token
	refreshToken, refreshPayload, err := s.tokenMake.CreateToken(user.Username, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "刷新令牌颁发失败, 内部错误: %s", err)
	}

	fmt.Println("refreshPayload.ExpiresAt.Time", refreshPayload.ExpiresAt.Time)
	sessions, err := s.store.CreateSessions(ctx, db.CreateSessionsParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
		IsBlocked:    false,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "内部错误: %s", err)
	}

	return &pb.LoginUserResponse{
		User: &pb.User{
			Username:         user.Username,
			FullName:         user.FullName,
			Email:            user.Email,
			PasswordChangeAt: timestamppb.New(user.PasswordChangedAt),
			CreateAt:         timestamppb.New(user.CreatedAt),
		},
		SessionId:             sessions.ID.String(),
		AccessToken:           token,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiresAt.Time),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiresAt.Time),
	}, nil
}
