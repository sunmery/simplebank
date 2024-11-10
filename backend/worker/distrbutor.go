package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

// 创建任务, 并通过redis队列把它们分发给worker

// TaskDistributor 为了方便接口mock测试, 编写成通用接口
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opt ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor 强制RedisTaskDistributor实现TaskDistributor接口
// 利用go编译器的自动类型检查, 如果没有实现该接口的方法编译器报错
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	// 创建一个新client
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
