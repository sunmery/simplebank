package validator

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	// 用户名格式: a-z A-Z 0-9 和中文字符
	isValidUsername = regexp.MustCompile(`^[a-zA-Z0-9\p{Han}]+$`).MatchString

	// 全名格式: a-z A-Z 0-9 _ 和空格 和中文字符
	isValidFullName = regexp.MustCompile(`^[a-zA-Z0-9_\s\p{Han}]+$`).MatchString
	// 密码格式:
	isValidPassword = regexp.MustCompile(`^[a-zA-Z0-9_@.,*]+$`).MatchString
)

func validateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("value: '%s' 应该在 %d 和 %d 之间", value, minLength, maxLength)
	}

	return nil
}

func ValidateUsername(value string) error {
	if err := validateString(value, 2, 16); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return fmt.Errorf("username: '%s' 应当为a-z A-Z 0-9与下划线的组合", value)
	}

	return nil
}
func ValidateFullName(value string) error {
	if err := validateString(value, 3, 50); err != nil {
		return err
	}
	if !isValidFullName(value) {
		return fmt.Errorf("full_name: '%s' 应当为a-z A-Z 0-9与下划线的组合", value)
	}

	return nil
}

func ValidatePassword(value string) error {
	if err := validateString(value, 6, 30); err != nil {
		return err
	}

	if !isValidPassword(value) {
		return fmt.Errorf("密码的格式: 可以使用的特殊字符: @,.*_ 并至少6字符长度, 并且不超过30字符长度")
	}
	return nil
}

func ValidateEmail(value string) error {
	if err := validateString(value, 3, 100); err != nil {
		return err
	}

	// 验证邮件地址
	if _, err := mail.ParseAddress(value); err != nil {
		return err
	}
	return nil
}
