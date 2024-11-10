package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

// GrpcLogger 中间件 记录gRPC日志
// 记录以下消息作为日志:
// 请求类型: grpc
// 请求方法: GET/PUT/PATCH/POST/DELETE
// 请求耗时: 毫秒
// 状态码: grpc的状态码, 数字类型
// 状态文本: 人性化的提示
func GrpcLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	// 计算请求的耗时
	startTime := time.Now()
	// 将请求转发给要处理的处理程序
	result, err := handler(ctx, req)
	// 当请求结束即停止耗时
	duration := time.Since(startTime)

	// 定义默认的错误
	statusCode := codes.Unknown
	// 从处理程序中获取错误状态码
	if s, ok := status.FromError(err); ok {
		statusCode = s.Code()
	}

	// 默认的输出是Info级别
	logger := log.Info()

	// 出现错误时, 将Info级别替换为错误级并带上错误消息
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.Type("context", struct {
		UserId    string
		Username  string
		IpAddr    string
		UserAgent string
	}{
		UserId:    "",
		Username:  "",
		IpAddr:    "",
		UserAgent: "",
	}).
		Str("environment", "").
		Str("service", "").
		Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Str("transactionId", "").
		Dur("duration", duration).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Msg("received gRPC request")

	return result, err
}

// ResponseRecorder 记录器
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// WriteHeader 重写http包的WriteHeader以获得状态码
func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// Write 重写http包的Write以获得响应主体
func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(rec.Body)
}

// HttpLogger 中间件 记录HTTP日志
// 记录以下消息作为日志:
// 请求类型: http
// 请求方法: GET/PUT/PATCH/POST/DELETE
// 路径: HTTP请求路径
// 请求耗时: 毫秒
// 状态码: HTTP状态码, 如果是错误妈, 则会额外添加body主体,显示错误详细信息
// 状态文本: HTTP状态文本
func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		// 由于HTTP没有返回值, 将重写http的响应以存储状态码与状态文本到记录器中
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}

		// 将新的日志记录器代替原有的响应器
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		// 将错误的响应写到body主体中
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.Str("protocol", "http").
			Str("method", req.Method).
			Str("url", req.RequestURI).
			Type("headers", struct {
				ContentType string
				Accept      string
			}{
				ContentType: req.Header.Get("Content-Type"),
				Accept:      req.Header.Get("Accept"),
			}).
			Dur("duration", duration).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Msg("received HTTP request")
	})
}
