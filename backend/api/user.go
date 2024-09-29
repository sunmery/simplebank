package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"simple_bank/pkg"

	"github.com/gin-gonic/gin"

	db "simple_bank/db/sqlc"
)

func (s *Server) CreateUser(ctx *gin.Context) {
	type CreateUserRequest struct {
		Username string `json:"username" binding:"required"`
		FullName string `json:"fullName" binding:"required"`
		Password string `json:"password" binding:"required,gte=6"`
		Email    string `json:"email" binding:"required,email"`
	}
	type CreateUserResponse struct {
		Username string `json:"username" binding:"required"`
		FullName string `json:"fullName" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	password, err := pkg.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		FullName:       req.FullName,
		HashedPassword: password,
		Email:          req.Email,
	}

	user, createErr := s.store.CreateUser(ctx, arg)
	if createErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(createErr, &pgErr) {
			switch pgErr.Code {
			case "23505":
				ctx.JSON(http.StatusForbidden, gin.H{
					"message": pgErr.Message,
					"code":    pgErr.Code,
					"body":    "用户名已存在",
				})
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := CreateUserResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (s *Server) GetUser(ctx *gin.Context) {
	type GetUserRequest struct {
		Username string `json:"username" binding:"required"`
	}

	var req GetUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, queryErr := s.store.GetUser(ctx, req.Username)
	if queryErr != nil {
		if errors.Is(queryErr, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(queryErr))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(queryErr))
		return
	}
	rsp := db.Users{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (s *Server) loginUser(ctx *gin.Context) {
	type loginUserRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,gte=6"`
	}

	type UserResponse struct {
		Username          string    `json:"username"`
		FullName          string    `json:"fullName"`
		Email             string    `json:"email"`
		PasswordChangedAt time.Time `json:"passwordChangedAt"`
		CreatedAt         time.Time `json:"createdAt"`
		UpdatedAt         time.Time `json:"updatedAt"`
	}

	type loginUserResponse struct {
		AccessToken string       `json:"accessToken"`
		User        UserResponse `json:"user"`
	}

	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// 查询客户端传递的username参数
	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		// 查不到
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		// 其它错误
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// 检查密码与hash之后的密码是否匹配
	if checkErr := pkg.CheckHashedPassword(req.Password, user.HashedPassword); checkErr != nil {
		// 401
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "密码错误",
		})
		return
	}

	// 颁发token
	token, createErr := s.tokenMake.CreateToken(user.Username, s.config.AccessTokenDuration)
	if createErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := UserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, loginUserResponse{
		AccessToken: token,
		User:        rsp,
	})
}
