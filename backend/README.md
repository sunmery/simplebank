## Simple Bank

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
