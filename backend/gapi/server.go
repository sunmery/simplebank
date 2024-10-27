package gapi

import (
	"fmt"
	"simple_bank/config"
	"simple_bank/pb"

	"simple_bank/pkg/token"

	"github.com/gin-gonic/gin"
	db "simple_bank/db/sqlc"
)

type Server struct {
	// 提供所有的函数调用,但是返回错误, 主要为了向前兼容
	// 可以并行处理多个RPC,而不会互相阻塞
	pb.UnimplementedCreateUserServiceServer
	config    *config.Config
	store     db.Store
	tokenMake token.Maker
	router    *gin.Engine
}

func NewServer(config *config.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("error creating token maker: %w", err)
	}

	server := &Server{
		config:    config,
		store:     store,
		tokenMake: tokenMaker,
	}

	return server, nil
}

// Start 启动
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}
