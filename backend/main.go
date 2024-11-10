package main

import (
	"context"
	"fmt"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"simple_bank/mail"
	"simple_bank/worker"

	"net"
	"net/http"
	"simple_bank/config"
	"simple_bank/gapi"
	"simple_bank/pb"

	"simple_bank/api"

	"github.com/jackc/pgx/v5/pgxpool"
	db "simple_bank/db/sqlc"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Panic().Err(err).Msg("无法读取环境变量")
	}
	log.Info().Str("DB_SOURCE", cfg.DBSource)
	log.Info().Str("RedisAddress", cfg.RedisAddress)
	// 开发模式启用人类更容易阅读的日志格式
	// 生产模式启用JSON包裹的日志以导出到日志存储后端
	switch cfg.ENVIRONMENT {
	case "development":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	case "production":
	}

	conn, newDBErr := pgxpool.New(context.Background(), cfg.DBSource)
	if newDBErr != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v", err))
	}

	store := db.NewStore(conn)

	// 连接redis
	// 连接redis的参数, 生产请添加TLS与账户密码
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddress,
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
	}
	// 创建新的任务分发器
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	// 异步任务处理器
	go runTaskProcessor(cfg, redisOpt, store)

	// 支持http请求的grpc网关
	go runGatewayServer(cfg, store, taskDistributor)
	runGrpcServer(cfg, store, taskDistributor)

	// http服务
	// runGinServer(cfg, store)
}

// 运行任务处理器
func runTaskProcessor(conf *config.Config, redis asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGMailSender(conf.EmailSenderName, conf.EmailSenderAddress, conf.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(mailer, redis, store)
	log.Info().Msg("运行任务处理器")

	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("无法启动任务处理器")
	}
}

// Grpc服务
func runGrpcServer(cfg *config.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	// rpc服务
	server, err := gapi.NewServer(cfg, store, taskDistributor)
	if err != nil {
		log.Error().Err(err)
	}

	// 一元拦截器, 拦截grpc请求并写入日志
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	// 创建grpc服务实例
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterUserServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	// 监听端口
	listen, lisErr := net.Listen("tcp", cfg.GRPCServerAddress)
	if lisErr != nil {
		log.Error().Err(err)
	}

	log.Info().Msgf("gRPC server listening on: %s", listen.Addr().String())

	// 启动grpc服务
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Error().Err(err)
	}
}

// Gateway服务. 编写grpc服务, 可以接收客户端的HTTP请求. 在进程内翻译grpc为http的响应并返回
func runGatewayServer(cfg *config.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	// rpc服务
	server, err := gapi.NewServer(cfg, store, taskDistributor)
	if err != nil {
		log.Error().Err(err)
	}

	// 进程内翻译, 仅支持 一元rpc, 即单个请求与单个响应
	jsonOption := gwruntime.WithMarshalerOption(gwruntime.MIMEWildcard, &gwruntime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := gwruntime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 调用grpc-gateway生成的注册服务
	err = pb.RegisterUserServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Error().Err(err)
	}

	//  创建多路复用器
	mux := http.NewServeMux()
	// 路由到grpc服务
	// mux.Handle("/", middleware.GrpcCORS(grpcMux))
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	// 监听端口
	listen, lisErr := net.Listen("tcp", cfg.HTTPServerAddress)
	if lisErr != nil {
		log.Error().Err(err)
	}

	log.Info().Msgf("HTTP server listening on: %s", listen.Addr().String())
	// 将日志作为中间件, 返回一个HTTP中间件处理的新HTTP处理器
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listen, handler)
	if err != nil {
		log.Error().Err(err)
	}
}

// HTTP 服务
func runGinServer(cfg *config.Config, store db.Store) {
	server, newServerErr := api.NewServer(cfg, store)
	if newServerErr != nil {
		log.Error().Err(newServerErr)
	}

	err := server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Error().Err(err)
	}
}
