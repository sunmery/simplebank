package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "simple_bank/db/sqlc"
	"simple_bank/pkg"
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

	verifyEmail, err := processor.Store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: pkg.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("无法创建电子邮件: %w", err)
	}

	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:30001/v1/verify_email?email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	html := fmt.Sprintf(`Hello %s
	Thank you for registering with us! </br>
	Please <a href="%s">click here</a> to verify your email address
	`, user.Username, verifyUrl)
	to := []string{user.Email}
	if processor.mailer.SendEmail(subject, html, to, nil, nil, nil) != nil {
		return fmt.Errorf("发送验证邮件失败: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("任务已处理")

	return nil
}
