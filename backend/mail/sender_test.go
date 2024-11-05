package mail

import (
	"github.com/stretchr/testify/require"
	"simple_bank/config"
	"testing"
)

func TestSendWithGmail(t *testing.T) {
	// 在CI跳过该测试邮件
	if testing.Short() {
		t.Skip("Skip, please run without -short")
	}

	conf, err := config.LoadConfig("../")
	if err != nil {
		require.NoError(t, err)
	}
	sender := NewGMailSender(conf.EmailSenderName, conf.EmailSenderAddress, conf.EmailSenderPassword)
	subject := "A Test email"
	context := "Hello Lisa this is test email"
	to := []string{"mail@mandala.chat"}
	// cc := []string{"xiconz@qq.com"}
	// bcc := []string{"xiconz@qq.com"}
	// attachFiles := []string{"../main.go"}

	err = sender.SendEmail(
		subject,
		context,
		to,
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
}
