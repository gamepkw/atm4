package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"main/domain"

	"github.com/go-redis/redis"
)

type mysqlTransactionRepository struct {
	conn  *sql.DB
	redis *redis.Client
}

// NewMysqlTransactionRepository will create an object that represent the transaction.Repository interface
func NewMysqlTransactionRepository(conn *sql.DB, redis *redis.Client) domain.TransactionRepository {
	return &mysqlTransactionRepository{
		conn:  conn,
		redis: redis,
	}
}

// func (m *mysqlTransactionRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.Transaction, err error) {
// 	rows, err := m.Conn.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		logrus.Error(err)
// 		return nil, err
// 	}

// 	defer func() {
// 		errRow := rows.Close()
// 		if errRow != nil {
// 			logrus.Error(errRow)
// 		}
// 	}()

// 	result = make([]domain.Transaction, 0)

// 	for rows.Next() {
// 		t := domain.Transaction{}
// 		acc := domain.Account{}
// 		accountID := acc.ID
// 		accountName := acc.OwnerName
// 		accountBalance := acc.Balance
// 		accountStatus := acc.Status
// 		accountCreatedAt := acc.CreatedAt
// 		accountUpdatedAt := acc.UpdatedAt
// 		accountIsdeleted := acc.IsDeleted
// 		err = rows.Scan(
// 			&t.TID,
// 			&t.Amount,
// 			&t.TrType,
// 			&t.CreatedBy,
// 			&t.ResipientID,
// 			// &t.CreatedAt,
// 		)

// 		if err != nil {
// 			logrus.Error(err)
// 			return nil, err
// 		}
// 		t.CreatedBy = domain.Account{
// 			ID:        accountID,
// 			OwnerName: accountName,
// 			Balance:   accountBalance,
// 			Status:    accountStatus,
// 			CreatedAt: accountCreatedAt,
// 			UpdatedAt: accountUpdatedAt,
// 			IsDeleted: accountIsdeleted,
// 		}

// 		t.ResipientID = domain.Account{
// 			ID:        accountID,
// 			OwnerName: accountName,
// 			Balance:   accountBalance,
// 			Status:    accountStatus,
// 			CreatedAt: accountCreatedAt,
// 			UpdatedAt: accountUpdatedAt,
// 			IsDeleted: accountIsdeleted,
// 		}

// 		result = append(result, t)

// 	}

// 	return result, nil
// }

// func (m *mysqlTransactionRepository) GetAllTransaction(ctx context.Context, cursor string, num int64) (res []domain.Transaction, nextCursor string, err error) {
// 	// query := `SELECT tid,name,content, author_tid, updated_at, created_at
// 	// 					FROM transaction WHERE created_at > ? ORDER BY created_at LIMIT ? `

// 	query := `SELECT * FROM atm.transaction WHERE created_at > ? ORDER BY created_at LIMIT ?`

// 	decodedCursor, err := repository.DecodeCursor(cursor)
// 	if err != nil && cursor != "" {
// 		return nil, "", domain.ErrBadParamInput
// 	}

// 	res, err = m.fetch(ctx, query, decodedCursor, num)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	if len(res) == int(num) {
// 		nextCursor = repository.EncodeCursor(res[len(res)-1].CreatedAt)
// 	}

// 	return
// }

// func (m *mysqlTransactionRepository) GetTransactionByTID(ctx context.Context, tid int64) (res domain.Transaction, err error) {
// 	query := `SELECT * FROM atm.transaction WHERE tid = ?`

// 	list, err := m.fetch(ctx, query, tid)
// 	if err != nil {
// 		return domain.Transaction{}, err
// 	}

// 	if len(list) > 0 {
// 		res = list[0]
// 	} else {
// 		return res, domain.ErrNotFound
// 	}

// 	return
// }

func (m *mysqlTransactionRepository) CreateTransaction(ctx context.Context, tr *domain.Transaction) (err error) {
	// query := `INSERT atm.transaction SET type=? , amount=? ,created_by=?, created_at=?`
	query := `
			INSERT INTO banking.transactions 
			SET amount=?, type=?, fee=?, total_amount=?, submitted_at=?, created_at=? , account=?, receiver=?
	`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	tr.CreatedAt = time.Now()

	res, err := stmt.ExecContext(ctx, tr.Amount, tr.Type, tr.Fee, tr.Total, tr.SubmittedAt, tr.CreatedAt, tr.Account.AccountNo, tr.Receiver.AccountNo)
	if err != nil {
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	tr.Id = lastID

	return nil
}

func (m *mysqlTransactionRepository) CreateScheduledTransaction(ctx context.Context, tr *domain.ScheduledTransaction) (err error) {
	query := `
			INSERT INTO banking.scheduled_transactions 
			SET amount=?, type=?, account=?, receiver=?, submitted_at=?, created_at=? , scheduled_execution_at =?, updated_at=?
	`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	tr.CreatedAt = time.Now()
	tr.UpdatedAt = time.Now()

	res, err := stmt.ExecContext(ctx, tr.Amount, tr.Type, tr.Account.AccountNo, tr.Receiver.AccountNo, tr.SubmittedAt, tr.CreatedAt, tr.ScheduledExecutionAt, tr.UpdatedAt)
	if err != nil {
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	tr.Id = lastID

	return nil
}

func (m *mysqlTransactionRepository) GetScheduledTransaction(ctx context.Context, fetchTime time.Time) (transactions []domain.ScheduledTransaction, err error) {

	query := `
			SELECT * FROM scheduled_transactions WHERE (scheduled_execution_at = ? AND status = 'unprocessed')
	`

	rows, err := m.conn.Query(query, fetchTime)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		transaction := domain.ScheduledTransaction{}
		err := rows.Scan(
			&transaction.Id,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Account.AccountNo,
			&transaction.Receiver.AccountNo,
			&transaction.Status,
			&transaction.SubmittedAt,
			&transaction.CreatedAt,
			&transaction.ScheduledExecutionAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		transactions = append(transactions, transaction)

		transaction.Status = "processing"
		m.UpdateScheduledTransaction(ctx, transaction)

	}

	return
}

func (m *mysqlTransactionRepository) UpdateScheduledTransaction(ctx context.Context, tr domain.ScheduledTransaction) (err error) {
	query := `UPDATE banking.scheduled_transactions set status=?, updated_at=? WHERE id = ?`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	tr.UpdatedAt = time.Now()

	_, err = stmt.ExecContext(ctx, tr.Status, tr.UpdatedAt, tr.Id)
	if err != nil {
		return
	}
	return
}

// func (m *mysqlAccountRepository) SetMaximumTransferAmountPerDayInRedis(ctx context.Context, id uint64) error {
// 	// Check if the user exists in Redis cache
// 	cacheKey := fmt.Sprintf("accountdailylimit:%d", id)

// 	MaxTransferAmountPerDay := 100000

// 	err := m.redis.Set(cacheKey, MaxTransferAmountPerDay, 0).Err()
// 	if err != nil {
// 		return err
// 	}
// 	return err

// }

// func (m *mysqlAccountRepository) UpdateMaximumTransferAmountPerDayInRedis(ctx context.Context, id uint64) error {
// 	// Check if the user exists in Redis cache
// 	cacheKey := fmt.Sprintf("accountdailylimit:%d", id)

// 	RemainTransferAmountPerDay

// 	err := m.redis.Set(cacheKey, RemainTransferAmountPerDay, 0).Err()
// 	if err != nil {
// 		return err
// 	}
// 	return err

// }

func (m *mysqlTransactionRepository) SetTransferAmountPerDayInRedis(ctx context.Context, tr *domain.Transaction) error {
	// Check if the user exists in Redis cache
	cacheKey := fmt.Sprintf("daily_transaction_%s", tr.Account.AccountNo)
	cachedAmount, err := m.redis.Get(cacheKey).Result()
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	expirationTime := nextMidnight
	durationUntilExpiration := expirationTime.Sub(now)

	if err == redis.Nil {
		err = m.redis.Set(cacheKey, tr.Amount, durationUntilExpiration).Err()
		if err != nil {
			return nil
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("error parsing user from cache: %v", err)
	} else {
		amount := cachedAmount
		floatamount, _ := strconv.ParseFloat(amount, 64)
		floatamount += tr.Amount
		err = m.redis.Set(cacheKey, floatamount, 0).Err()
		if err != nil {
			return fmt.Errorf("error parsing user from cache: %v", err)
		}
		return nil
	}

}

func (m *mysqlTransactionRepository) MigrateTransactionHistory(ctx context.Context) (err error) {
	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")

	log.Printf("Starting migration for date: %s", currentDate)

	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	log.Printf("Counting rows to be migrated...")

	select_query := `
					SELECT COUNT(*) FROM banking.transactions 
					WHERE created_at < ?;
	`

	stmt, err := m.conn.PrepareContext(ctx, select_query)
	if err != nil {
		return err
	}

	var countRows int64
	err = stmt.QueryRowContext(ctx, currentDate).Scan(&countRows)
	if err != nil {
		return err
	}

	migrate_query := `
					INSERT INTO banking.transactions_history
					SELECT * FROM banking.transactions
					WHERE created_at < ?;
	`

	stmt, err = m.conn.PrepareContext(ctx, migrate_query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, currentDate)
	if err != nil {
		return err
	}

	insertRowAffect, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if countRows != insertRowAffect {
		tx.Rollback()
		return fmt.Errorf("number of rows affected does not match")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	delete_query := `
					DELETE FROM banking.transactions
					WHERE created_at < ?;
	`

	stmt, err = m.conn.PrepareContext(ctx, delete_query)
	if err != nil {
		return
	}

	res, err = stmt.ExecContext(ctx, currentDate)
	if err != nil {
		return err
	}

	deleteRowAffect, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if deleteRowAffect != insertRowAffect {
		tx.Rollback()
		return fmt.Errorf("number of rows affected in delete and migrate operations do not match")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Migration completed successfully")

	return nil
}
