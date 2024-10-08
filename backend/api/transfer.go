package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"simple_bank/constants"
	"simple_bank/pkg/token"

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
	fromAccount, valid := s.validateCurrent(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	// 比较账户的Owner与gin ctx token的payload的Username
	// 如果一致, 说明登录的用户是该账户的拥有者
	// 不一致抛出异常
	payload := ctx.MustGet(constants.AuthorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != payload.Username {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorResponse(errors.New("登录的用户非该账户的拥有者"))})
		return
	}

	_, valid = s.validateCurrent(ctx, req.ToAccountID, req.Currency)
	if valid {
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
func (s *Server) validateCurrent(ctx *gin.Context, accountID int64, currency string) (db.Accounts, bool) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		// 数据库不存在的情况
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		// 服务器内部错误
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	if currency != account.Currency {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "转账时请使用相同的货币类型",
			"error":   fmt.Sprintf("用户ID'%d' 的货币类型不匹配: '%s' vs '%s'", accountID, currency, account.Currency),
		})
		return account, false
	}
	return account, true
}
