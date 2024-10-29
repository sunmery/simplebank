package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"simple_bank/constants"

	db "simple_bank/db/sqlc"
)

//	处理器
//
// 从Redis队列获取任务并处理
type TaskProcessor interface {
	Start() error
	ProcessTaskVerifyEmail(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor Redis任务处理器
// Store 在需要时操作数据库
type RedisTaskProcessor struct {
	Server *asynq.Server
	Store  db.Store
}

// NewRedisTaskProcessor 任务处理器实例
func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		// 队列优先级
		Queues: map[string]int{
			constants.QueueCritical: 6,
			constants.QueueDefault:  3,
			constants.QueueLow:      1,
		},
		// 处理异步任务的错误, 以符合log日志的格式
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().
				Err(err).
				Str("type", task.Type()).
				Bytes("payload", task.Payload()).
				Msg("处理任务失败")
		}),
		Logger: NewLogger(),
	})

	return &RedisTaskProcessor{
		Server: server,
		Store:  store,
	}
}

// Start 在启动异步服务器之前将任务队列函数进行注册
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	// 注册任务队列
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskVerifyEmail)

	err := processor.Server.Start(mux)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

// ProcessTaskVerifyEmail 处理验证邮件
// asynq已经处理了redis中拉取任务部分,并通过此处理函数的任务参数将其提供给后台的worker进行处理
func (processor *RedisTaskProcessor) ProcessTaskVerifyEmail(ctx context.Context, task *asynq.Task) error {
	// 解析任务并获取其有效负载
	var payload PayloadSendVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("unmarshal payload verify email payload: %w", err)
	}

	// 查询用户
	user, err := processor.Store.GetUser(ctx, payload.Username)
	if err != nil {
		// 如果查不到此用户, 那么跳过 重试任务
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found %w", asynq.SkipRetry)
		}
		return fmt.Errorf("get user: %w", err)
	}
	// TODO: send email to user
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("任务已处理")

	return nil
}
