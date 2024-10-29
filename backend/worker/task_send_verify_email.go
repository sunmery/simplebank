package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// PayloadSendVerifyEmail 发送验证邮件 结构
type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

const TaskSendVerifyEmail = "task:send_verify_email"

// DistributeTaskSendVerifyEmail 发送验证邮件的任务创建和分配
func (distribute *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode payload error: %w", err)
	}
	// 创建一个新任务
	task := asynq.NewTask(TaskSendVerifyEmail, bytes, opts...)
	// 将任务 发送到redis队列
	taskInfo, err := distribute.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("未能将任务进入队列: %w", err)
	}

	// 队列类型
	// 队列名称
	// 最大重试次数
	// 荷载
	log.Info().
		Str("type", task.Type()).
		Str("queue", taskInfo.Queue).
		Int("max_retry", taskInfo.MaxRetry).
		Bytes("payload", taskInfo.Payload).
		Msg("任务已进入队列")

	return nil
}
