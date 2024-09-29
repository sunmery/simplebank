package pkg

import (
	"fmt"
	"math/rand"
	"simple_bank/constants"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomInt 在min, max的范围随机生成数字
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString 生成n个字符
func RandomString(n int) string {
	// 声明一个字符串的缓冲区, 对于逐个字节插入非常的高效
	var sb strings.Builder
	// 获取字符表的长度, 减少for多次访问len()函数
	k := len(alphabet)

	// 获取n个字符
	for i := 0; i < n; i++ {
		// 从字符表中随机选取一个字符
		c := alphabet[rand.Intn(k)]
		// 写入到缓冲区对象
		sb.WriteByte(c)
	}
	// 将buffer流转成字符串
	return sb.String()
}

// RandomCurrency 随机生成货币类型
func RandomCurrency() string {
	currency := []string{constants.CAD, constants.USD, constants.CNY}
	return currency[rand.Intn(len(currency))]
}

func RandomEmail(len int) string {
	return fmt.Sprintf("%s@email.com", RandomString(len))
}
