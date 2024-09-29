package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransfersParams) (TransfersTxResult, error)
}

type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx 通用的事务方法, 通过外部传递函数作为事务的运行内容
func (s *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// 开始一个事务, 如sql的begin
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	// 事务的运行内容
	query := New(tx)
	err = fn(query)
	// 如果事务发生错误
	if err != nil {
		// 如果回滚发生错误, 那么合并两个错误为一个错误返回回去
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err is: '%s', rellback err is: '%s'", err, rbErr)
		}
		// 返回事务错误
		return err
	}

	// 没有错误则提交该事务
	return tx.Commit(ctx)
}

type TransfersParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransfersTxResult struct {
	Transfer    Transfers `json:"transfer"`
	FromAccount Accounts  `json:"from_account"`
	ToAccount   Accounts  `json:"to_account"`
	FromEntry   Entries   `json:"from_entry"`
	ToEntry     Entries   `json:"to_entry"`
}

// TransferTx 转账方法
// 1. 转账表记录一条数据, 是谁向谁发送了转账记录
// 2. 条目表记录一条数据, 记录用户转出的金额
// 3. 条目表记录一条数据, 记录用户转入的金额
// 4. 账户表记录一条数据, 记录账户的转出
// 5. 账户表记录一条数据, 记录账户的转入
func (s *SQLStore) TransferTx(ctx context.Context, arg TransfersParams) (TransfersTxResult, error) {
	var result TransfersTxResult

	// 执行转账事务
	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		// 转账表记录一条数据, 是谁向谁发送了转账记录
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// 条目表记录一条数据, 记录用户支出的金额
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// 条目表记录一条数据, 记录用户收入的金额
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// 排序代码的执行顺序, 让sql语句按照顺序执行,
		// 否则在并发时会引发这种情况: 事务1需要修改事务2中的行时,需要等待事务2提交或回滚, 事务2也操作了事务1中的行也要等待事务1提交或回滚 造成死锁
		// 因为在遇到UPDATE更新时, 数据库默认自动给该事务添加行级排它锁,
		// 它会阻止其它事务对该行的修改操作(但不影响查询)
		if arg.FromAccountID < arg.ToAccountID {
			// 直接更新余额, 将查询该账户获取该账号的id与账号更新余额合并为一个方法
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}
		return err
	})

	return result, err
}

// 不直接写一个amount是因为后续可能有汇率或者手续费相关的东西
func addMoney(
	ctx context.Context,
	q *Queries,
	account1ID int64,
	amount1 int64,
	account2ID int64,
	amount2 int64,
) (account1 Accounts, account2 Accounts, err error) {
	account1, err = q.AddAccountBalancer(ctx, AddAccountBalancerParams{
		Amount: amount1,
		ID:     account1ID,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalancer(ctx, AddAccountBalancerParams{
		Amount: amount2,
		ID:     account2ID,
	})
	// 命名的返回值会自动填充, 以下的return就相当于account1, account2, err
	return
}
