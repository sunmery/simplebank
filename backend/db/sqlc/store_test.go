package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Printf("交易前的余额: account1 :%d, account2: %d \n", account1.Balance, account2.Balance)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransfersTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), TransfersParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// 创建了一个空的 map，用来存储已经出现过的整数 k，防止重复。
	existed := make(map[int]bool)

	// 检查结果

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// 检查转账操作
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// 检查条目操作
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		require.WithinDuration(t, toEntry.CreatedAt, account2.CreatedAt, time.Second)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// 检查账户
		fromAccount := result.FromAccount

		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// 检查账户余额
		// 计算两个账户余额的差值，并将差值除以某个单位（amount）后得到一个倍数 k，
		// 然后对这个 k 值进行断言验证，确保差值合理、范围合法，且不重复出现
		fmt.Printf("事务交易后的余额: account1 :%d, account2: %d \n", fromAccount.Balance, toAccount.Balance)

		// 计算差值
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance

		// 如果事务正常运行,则两个diff是相同的
		require.Equal(t, diff1, diff2)

		// diff1的值应该是正数
		require.True(t, diff1 > 0)
		// 差异是可取余的, 它们的倍数成正比, 第一次交易是1倍, 第二次交易是2倍, 例如第一次是1 * amount,第二次为2 * amount
		// 等于0则代表正确的交易
		require.True(t, diff1%amount == 0)

		// 断言 diff1 必须大于 0，确保差值是正数。
		k := int(diff1 / amount)
		// 确保符合go携程数符合定义的数与事务运行的数量相同
		require.True(t, k >= 1 && k <= n)
		// 判断k是否存在与existed中, 存在则报错
		require.NotContains(t, existed, k)
		// 现有的map不应该包含k
		// 将 k 添加到 existed map 中，表示 k 已经被处理过，防止重复处理相同的 k
		existed[k] = true
	}

	// 检查账户最终更新的余额
	// 从数据库获取更新后的账户1
	updateAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.ID, updateAccount1.ID)
	// 与原始的余额进行对比, account1是支出, 则余额应该是原始余额支出后的余额

	updateAccount2, err2 := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err2)

	// fmt.Printf("交易后的余额: account1 :%d, account2: %d \n", account1.Balance, account2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updateAccount1.Balance)
	// 与原始的余额进行对比, account1是收入, 则余额应该是原始余额收入后的余额
	require.Equal(t, account2.Balance+int64(n)*amount, updateAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Printf("交易前的余额: account1 :%d, account2: %d \n", account1.Balance, account2.Balance)

	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		// 奇数的携程将account1.ID的账户的余额转到account2.ID的账户
		// 偶数的携程将account2.ID的账户的余额转到account1.ID的账户
		// 它们最终的结果如果是相同的,则说明并发安全
		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransfersParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	// 检查结果
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// 检查账户最终更新的余额
	// 从数据库获取更新后的账户1
	updateAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.ID, updateAccount1.ID)
	updateAccount2, err2 := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err2)

	// fmt.Printf("交易后的余额: account1 :%d, account2: %d \n", account1.Balance, account2.Balance)
	// 与原始的余额进行对比, 应该是相同的余额, 无变化
	require.Equal(t, account1.Balance, updateAccount1.Balance)
	// 与原始的余额进行对比, 应该是相同的余额, 无变化
	require.Equal(t, account2.Balance, updateAccount2.Balance)
}
