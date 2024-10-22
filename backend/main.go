package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"simple_bank/gapi"
	"simple_bank/pb"

	"simple_bank/config"

	"simple_bank/api"

	"github.com/jackc/pgx/v5/pgxpool"
	db "simple_bank/db/sqlc"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	conn, newDBErr := pgxpool.New(context.Background(), cfg.DBSource)
	if newDBErr != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v", err))
	}

	store := db.NewStore(conn)

	runGrpcServer(cfg, store)
}

// Grpc服务
func runGrpcServer(cfg *config.Config, store db.Store) {
	// rpc服务
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		panic(fmt.Sprintf("Unable to create server: %v", err))
	}

	// 创建grpc服务实例
	grpcServer := grpc.NewServer()
	pb.RegisterCreateUserServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	// 监听端口
	listen, lisErr := net.Listen("tcp", cfg.GRPCServerAddress)
	if lisErr != nil {
		panic(fmt.Sprintf("Unable to create server port: %v", lisErr))
	}

	log.Printf("gRPC server listening on: %s", listen.Addr().String())

	// 启动grpc服务
	err = grpcServer.Serve(listen)
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %v", err))
	}

}

// Gateway服务. 接收HTTP与GRPC请求
func runGatewayServer(cfg *config.Config, store db.Store) {
	// rpc服务
	server, err := gapi.NewServer(cfg, store)
	if err != nil {
		panic(fmt.Sprintf("Unable to create server: %v", err))
	}

	// 进程内翻译, 仅支持 一元rpc, 即单个请求与单个响应
	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 调用grpc-gateway生成的注册服务
	err = pb.RegisterCreateUserServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		panic(fmt.Sprintf("Unable to create server: %v", err))
	}

	// 监听端口
	listen, lisErr := net.Listen("tcp", cfg.HTTPServerAddress)
	if lisErr != nil {
		panic(fmt.Sprintf("Unable to create server port: %v", lisErr))
	}

	log.Printf("gRPC server listening on: %s", listen.Addr().String())

	// 启动grpc服务
	err = grpcServer.Serve(listen)
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %v", err))
	}
}

// HTTP 服务
func runGinServer(cfg *config.Config, store db.Store) {
	server, newServerErr := api.NewServer(cfg, store)
	if newServerErr != nil {
		panic(fmt.Sprintf("Unable to create server: %v", newServerErr))
	}

	err := server.Start(cfg.HTTPServerAddress)
	if err != nil {
		panic(fmt.Sprintf("Unable to start server: %v", err))
	}
}
