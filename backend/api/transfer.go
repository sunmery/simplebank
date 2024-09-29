package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "simple_bank/db/sqlc"

	"github.com/gin-gonic/gin"
)

// 创建转账记录
func (s *Server) createTransfer(ctx *gin.Context) {
	type CreateTransferRequest struct {
		FromAccountID int64  `json:"fromAccountID" binding:"required"`
		ToAccountID   int64  `json:"toAccountID" binding:"required"`
		Amount        int64  `json:"amount" binding:"required,gte=0"`
		Currency      string `json:"currency" binding:"required,currency"`
	}

	var req CreateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorResponse(err)})
		return
	}

	// 创建转账记录时, 需要判断传入的,发起者的,接受者货币类型,三者一致才可以创建转账条目
	if !s.validateCurrent(ctx, req.FromAccountID, req.Currency) {
		return
	}
	if !s.validateCurrent(ctx, req.ToAccountID, req.Currency) {
		return
	}

	result, err := s.store.CreateTransfer(ctx, db.CreateTransferParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	})
	fmt.Printf("result:%v", result)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errorResponse(err)})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

// 验证货币类型
func (s *Server) validateCurrent(ctx *gin.Context, accountID int64, currency string) bool {
	toAccount, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		// 数据库不存在的情况
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		// 服务器内部错误
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}
	if currency != toAccount.Currency {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "转账时请使用相同的货币类型",
			"error":   fmt.Sprintf("用户ID'%d' 的货币类型不匹配: '%s' vs '%s'", accountID, currency, toAccount.Currency),
		})
		return false
	}
	return true
}
