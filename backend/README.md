## Simple Bank

### 错误处理

postgres错误码对照表: http://www.postgres.cn/docs/9.4/errcodes-appendix.html

### 发送电子邮件
https://www.nodeseek.com/post-183150-1

```go
package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"simple_bank/constants"
)

type EmailSender interface {
	// SendEmail subject主题, body内容, to收件人邮件地址, cc(Carbon Copy)抄送人邮件地址, bcc(Blind Carbon Copy)密件抄送人邮件地址 attachFiles附件
	SendEmail(subject, context string, to []string, cc []string, bcc []string, attachFiles []string) error
}

type GMailSender struct {
	name              string
	formEmailAddress  string
	formEmailPassword string
}

// NewGMailSender
// 发件人
// 发件人电子邮箱
// 发件人邮箱密码
func NewGMailSender(name string, formEmailAddress string, formEmailPassword string) EmailSender {
	return &GMailSender{
		name:              name,
		formEmailAddress:  formEmailAddress,
		formEmailPassword: formEmailPassword,
	}
}

func (sender GMailSender) SendEmail(
	subject string,
	context string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.formEmailAddress)
	e.Subject = subject
	e.HTML = []byte(context)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, file := range attachFiles {
		_, err := e.AttachFile(file)
		if err != nil {
			return fmt.Errorf("attach file %s failed: %w", file, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.formEmailAddress, sender.formEmailPassword, constants.GmailSMTPAuthHost)
	return e.Send(constants.GmailSMTPServerAddress, smtpAuth)
}

```

### 验证电子邮件
1. 生成随机密码到数据库库中
2. 将代码发送到用户注册的邮件中
3. 代码包含指向验证邮件的路由进行验证
4. 用户拥有该电子邮件则可以使用正确的密码访问页面

实现:
1. 用户表添加 is_email_verified 的布尔值非空类型, 表示用户是否通过验证邮件, 默认值为false
2. 创建验证表: verify_emails ,存储密码和有关发送给用户的电子邮件的所有信息
   1. id pk
   2. username varchar, 用户名 引用用户表的username
   3. email varchar, 用于跟踪用户的电子邮件地址, 可用于未来修改用户的邮件, 不唯一, 可以连续向用户发送多封验证邮件
   4. secret_code varchat not null, 随机生成的一次性密码
   5. is_used boolean 默认false, 是否使用, 标记是否被使用过, 安全性, 密码应该只能被使用一次, 不能被重复使用
   6. created_at timestamptz, 记录用户创建该验证邮件的时间
   7. expired_at timestamptz default now() + interval '15 minutes', 过期时间, 默认15分钟后, 何时过期, 过期则失效, 无法验证, 安全性, 减少泄露风险 
3. 创建verify_email.sql,编写CURD,
   1. 插入: 用户名, 邮件, 一次性密码
4. 将验证邮件加入到任务队列. 更新代码
4.1
`task_send_verify_email.go`
```go
verifyEmail, err := processor.Store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
    Username:   user.Username,
    Email:      user.Email,
    SecretCode: pkg.RandomString(32),
})
if err != nil {
    return fmt.Errorf("无法创建电子邮件: %w", err)
}
```
4.2. 在Redis任务处理器RedisTaskProcessor结构体添加mailer发送邮件的方法
```go
type RedisTaskProcessor struct {
   mailer mail.EmailSender
}
```
4.3 更新NewRedisTaskProcessor
```go
func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
   return &RedisTaskProcessor{
     Server: server,
     Store:  store,
     mailer: mailer,
}
```
4.4 完善发送验证邮件
```go
subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:30001/verify_email?id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	html := fmt.Sprintf(`Hello %s
	Thank you for registering with us! </br>
	Please <a href="%s">click here</a> to verify your email address
	`, user.Username, verifyUrl)
	to := []string{user.Email}
	if processor.mailer.SendEmail(subject, html, to, nil, nil, nil) != nil {
		return fmt.Errorf("发送验证邮件失败: %w", err)
	}
```
5. 更新main.go
```go
cfg, err := config.LoadConfig(".")
go runTaskProcessor(cfg, redisOpt, store)

func runTaskProcessor(conf *config.Config, redis asynq.RedisClientOpt, store db.Store) {
  mailer := mail.NewGMailSender(conf.EmailSenderName, conf.EmailSenderAddress, conf.EmailSenderPassword)
  taskProcessor := worker.NewRedisTaskProcessor(mailer, redis, store)
  log.Info().Msg("运行任务处理器")

  err := taskProcessor.Start()
     if err != nil {
     log.Fatal().Err(err).Msg("无法启动任务处理器")
	 }
}
```
6. 验证电子邮件, 用户点击验证按钮时, 将根据url的id查询该sector_code的值与数据库的值对比是否一致且没有过期时, 则标记为已使用, 将用户表的is_verify_email修改为true
```protobuf
syntax = "proto3";

// 验证用户电子邮件

package simple_bank;

option go_package = "simple_bank/pb";

message VerifyEmailRequest {
   int64 email_id = 1;
   string secret_code = 2;
}

message VerifyEmailResponse {
   bool is_verified = 1;
}

service UserService {
   rpc VerifyEmail(VerifyEmailRequest) returns (VerifyEmailResponse) {
      option (google.api.http) = {
         get: "/v1/verify_email"
      };
      option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
         description: "使用此API验证用户邮件",
         summary: "验证电子邮件"
      };
   }
}
```
6.1 添加校验
validator.go
```go
func ValidateEmailId(value int64) error {
   if value <= 0 {
      return fmt.Errorf("email_id必须是正整数")
   }
   return nil
}

func ValidateEmailSecretCode(value string) error {
   return validateString(value, 32, 128)
}
```
6.2 校验验证邮件RPC参数
```go
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

```

7. 编写sql,verify_emails:
```sql
-- name: UpdateVerifyEmail :one
-- 更新is_used为已使用(TRUE),
-- 条件是一次性密码(secret_code)相同且没有使用过(is_used = FALSE)和在有效期内(expired_at > now())
UPDATE verify_emails
SET is_used = TRUE
WHERE id = @id
  AND secret_code = @secret_code
  AND is_used = FALSE
  AND expired_at > now()
RETURNING *;
```

8. 添加更新users功能的is_email_verified列:
```sql
-- name: UpdateUser :one
UPDATE users
SET
     username = coalesce(sqlc.narg(username), username),
     full_name = coalesce(sqlc.narg(full_name), full_name),
     hashed_password = coalesce(sqlc.narg(hashed_password), hashed_password),
     email = coalesce(sqlc.narg(email), email),
     is_email_verified = coalesce(sqlc.narg(is_email_verified), is_email_verified)
WHERE username = sqlc.arg(username)
RETURNING *;
```

9. tx_verify_email.go实现事务
```go
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

```
10. 添加到store:
```go
type Store interface {
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

```

调用, rpc_verify_email.go:
```go
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
```

编写测试
