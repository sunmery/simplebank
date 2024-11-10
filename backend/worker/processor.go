package worker

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"simple_bank/mail"

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
	mailer mail.EmailSender
}

// NewRedisTaskProcessor 任务处理器实例
func NewRedisTaskProcessor(mailer mail.EmailSender, redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
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
		mailer: mailer,
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
