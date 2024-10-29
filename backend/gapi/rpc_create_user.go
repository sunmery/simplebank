package gapi

import (
	"context"
	"database/sql"
	"errors"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"simple_bank/constants"
	db "simple_bank/db/sqlc"
	"simple_bank/pb"
	"simple_bank/pkg"
	"simple_bank/validator"
	"simple_bank/worker"
	"time"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 校验rpc参数
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidCounterargument(violations)
	}

	password, err := pkg.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "无法获取密码")
	}

	// 将数据库创建用户与用户发送验证邮件的任务绑定, 一起提交或回滚
	arg := db.CreateUserTxParams{
		// 创建用户
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			FullName:       req.GetFullName(),
			HashedPassword: password,
			Email:          req.GetEmail(),
		},
		// 发送验证邮件
		// 如果此函数失败则当前事务回滚
		AfterCreate: func(user db.Users) error {
			// 发送验证邮件
			// 创建用户应当与发送邮件任务队列一起成功或者失败, 应使用事务
			// 否则数据库创建用户成功, 但是异步任务失败, 客户端将收到内部错误, 但无法重试, 因为会创建相同用户名的重复记录
			taskPayload := &worker.PayloadSendVerifyEmail{Username: user.Username}
			opts := []asynq.Option{
				asynq.MaxRetry(10),                   // 最大重试次数
				asynq.ProcessIn(10 * time.Second),    // 延迟n单位后被处理器接收
				asynq.Queue(constants.QueueCritical), // 优先级, 例如critical: 关键
			}

			return s.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}
	txResult, err := s.store.CreateUserTx(ctx, arg)
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
			Username:         txResult.User.Username,
			FullName:         txResult.User.FullName,
			Email:            txResult.User.Email,
			PasswordChangeAt: timestamppb.New(txResult.User.PasswordChangedAt),
			CreateAt:         timestamppb.New(txResult.User.CreatedAt),
		},
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	// 校验登录参数
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidCounterargument(violations)
	}

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

	mtdt := s.extractMetadata(ctx)
	sessions, err := s.store.CreateSessions(ctx, db.CreateSessionsParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
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

// 校验注册用户RPC参数
func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	// 应为proto中定义的字段名, 即蛇形命名法的字段
	if err := validator.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}
	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	return
}

// 校验用户登录RPC参数
func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return
}
