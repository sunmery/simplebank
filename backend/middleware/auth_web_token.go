package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"simple_bank/constants"
	"simple_bank/pkg/token"
	"strings"
)

func AuthWebTokenMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取AuthorizationHeader头这个key的值
		header := ctx.GetHeader(constants.AuthorizationHeaderKey)
		if header == "" || len(header) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errors.New("未提供 authorization 标头"),
			})
			return
		}
		// 根据获取到的字段的值进行拆分为两个不同的切片元素
		fields := strings.Fields(header)
		// 判断切片是否合法
		if len(fields) != 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errors.New("授权标头格式无效"),
			})
		}

		// 判断切片是否为服务器支持的授权类型
		if strings.ToLower(fields[0]) != constants.AuthorizationHeaderType {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errors.New("服务器不支持的授权类型"),
			})
		}

		// 获取Authorization头的值
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		}
		ctx.Set(constants.AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
