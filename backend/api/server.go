package api

import (
	"fmt"
	"log"
	"simple_bank/middleware"

	"simple_bank/config"

	"simple_bank/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	db "simple_bank/db/sqlc"
)

type Server struct {
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

	server.setupRouter()

	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := validate.RegisterValidation("currency", validCurrency)
		if err != nil {
			log.Fatalf("error registering validation: %v", err)
		}
	}

	return server, nil
}

func (s *Server) setupRouter() {
	// 	gin.SetMode(gin.ReleaseMode)
	routes := gin.Default()
	routes.Use(middleware.Cors())

	// 创建单个用户
	routes.PUT("/users", s.CreateUser)

	// 用户登录
	routes.POST("/users/login", s.loginUser)

	// 查询单个用户
	authGroup := routes.Group("/").Use(middleware.AuthWebTokenMiddleware(s.tokenMake))

	authGroup.GET("/users", s.GetUser)

	// 创建单个账户
	authGroup.PUT("/accounts", s.createAccount)
	// 获取单个账户信息
	authGroup.GET("/accounts/:id", s.getAccount)
	// 获取账户列表信息
	authGroup.GET("/accounts", s.listAccount)

	// 创建转账记录
	authGroup.PUT("/transfers", s.createTransfer)

	s.router = routes
}

// Start 启动
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
