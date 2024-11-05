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
