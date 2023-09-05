package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"main/atm/repository"
	"main/domain"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type mysqlAccountRepository struct {
	conn  *sql.DB
	redis *redis.Client
}

// NewMysqlAccountRepository will create an object that represent the account.Repository interface
func NewMysqlAccountRepository(conn *sql.DB, redis *redis.Client) domain.AccountRepository {
	return &mysqlAccountRepository{
		conn:  conn,
		redis: redis,
	}
}

func (m *mysqlAccountRepository) getAllAccount(ctx context.Context, query string, args ...interface{}) (accounts []domain.Account, err error) {

	rows, err := m.conn.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	accounts = make([]domain.Account, 0)

	for rows.Next() {
		account := domain.Account{}

		err = rows.Scan(
			&account.AccountNo,
			&account.Uuid,
			&account.Name,
			&account.Email,
			&account.Tel,
			&account.Balance,
			&account.Bank,
			&account.Status,
			&account.IsClosed,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			logrus.Error(err)
			return accounts, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (m *mysqlAccountRepository) GetAllAccount(ctx context.Context, cursor string, num int64) (res []domain.Account, nextCursor string, err error) {
	query := `SELECT * FROM banking.accounts WHERE created_at > ? ORDER BY created_at LIMIT ? `

	decodedCursor, err := repository.DecodeCursor(cursor)
	if err != nil && cursor != "" {
		return nil, "", domain.ErrBadParamInput
	}

	res, err = m.getAllAccount(ctx, query, decodedCursor, num)
	if err != nil {
		return nil, "", err
	}

	if len(res) == int(num) {
		nextCursor = repository.EncodeCursor(*res[len(res)-1].CreatedAt)
	}

	return
}

func (m *mysqlAccountRepository) GetAccountByAccountNo(ctx context.Context, account_no string) (res *domain.Account, err error) {
	account, err := m.fetchAccountFromDatabase(ctx, account_no)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (m *mysqlAccountRepository) GetAllAccountByUuid(ctx context.Context, uuid string) (res *[]domain.Account, err error) {
	accounts, err := m.fetchAllAccountFromDatabaseByUuid(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

func (m *mysqlAccountRepository) GetCountAccountByStatus(ctx context.Context) (result map[string]int, err error) {
	query := `SELECT status_list.status, COALESCE(status_count.count, 0) AS count
	FROM (
		SELECT 'active' AS status
		UNION SELECT 'inactive'
		UNION SELECT 'fraud'
		UNION SELECT 'zero'
	) AS status_list
	LEFT JOIN (
		SELECT status, COUNT(*) AS count
		FROM banking.accounts
		WHERE status IN ('active', 'inactive', 'fraud', 'zero')
		GROUP BY status
	) AS status_count ON status_list.status = status_count.status`
	rows, err := m.conn.QueryContext(ctx, query)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	statusCounts := make(map[string]int)

	for rows.Next() {
		countAccount := domain.CountAccount{}

		err = rows.Scan(
			&countAccount.Status,
			&countAccount.Count,
		)

		if err != nil {
			return nil, err
		}

		statusCounts[countAccount.Status] = countAccount.Count
	}

	return statusCounts, nil
}

func (m *mysqlAccountRepository) GetAccountFromRedisByAccountNo(ctx context.Context, account_no string) (*domain.Account, error) {
	// Check if the user exists in Redis cache
	cacheKey := fmt.Sprintf("account_no: %s", account_no)
	cachedAccount, err := m.redis.Get(cacheKey).Result()

	if err == redis.Nil {
		// Cache miss: key does not exist in Redis
		account, err := m.fetchAccountFromDatabase(ctx, account_no)
		if err != nil {
			return nil, err
		}

		ttl := 0.0 * time.Second

		serializedAccount := m.serializeAccount(&account)
		err = m.redis.Set(cacheKey, serializedAccount, ttl).Err()
		if err != nil {
			return nil, err
		}
		return &account, nil
	} else if err != nil {
		return nil, fmt.Errorf("error parsing account from cache: %v", err)
	} else {
		// Cache hit: key exists in Redis, use the retrieved value
		account, err := m.parseAccountFromCache(cachedAccount)
		if err != nil {
			return nil, fmt.Errorf("error parsing account from cache: %v", err)
		}
		return account, nil
	}
}

func (m *mysqlAccountRepository) fetchAllAccountFromDatabaseByUuid(ctx context.Context, uuid string) (accounts []domain.Account, err error) {
	query := `SELECT * FROM banking.accounts WHERE uuid = ?`

	rows, err := m.conn.QueryContext(ctx, query, uuid)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	fmt.Println(uuid)

	accounts = make([]domain.Account, 0)

	for rows.Next() {
		account := domain.Account{}

		err = rows.Scan(
			&account.AccountNo,
			&account.Uuid,
			&account.Name,
			&account.Email,
			&account.Tel,
			&account.Balance,
			&account.Bank,
			&account.Status,
			&account.IsClosed,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			logrus.Error(err)
			return accounts, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (m *mysqlAccountRepository) fetchAccountFromDatabase(ctx context.Context, account_no string) (res domain.Account, err error) {

	query := `SELECT * FROM banking.accounts WHERE account_no = ?`

	list, err := m.getAllAccount(ctx, query, account_no)
	if err != nil {
		return domain.Account{}, err
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}

	return
}

func (m *mysqlAccountRepository) parseAccountFromCache(cachedContact string) (*domain.Account, error) {
	var account domain.Account
	err := json.Unmarshal([]byte(cachedContact), &account)
	if err != nil {
		return nil, fmt.Errorf("error parsing user from cache: %v", err)
	}
	return &account, nil
}

func (m *mysqlAccountRepository) serializeAccount(account *domain.Account) string {
	jsonData, _ := json.Marshal(account)
	return string(jsonData)
}

func (m *mysqlAccountRepository) RegisterAccount(ctx context.Context, a *domain.Account) (err error) {
	query := `INSERT banking.accounts SET account_no=?, uuid=?, name=? , email=? , tel=?, bank=? , created_at=? , updated_at=?`
	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	_, err = stmt.ExecContext(ctx, a.AccountNo, a.Uuid, a.Name, a.Email, a.Tel, a.Bank, time.Now(), time.Now())
	if err != nil {
		return
	}
	// lastID, err := res.LastInsertId()
	// if err != nil {
	// 	return
	// }
	// a.Uuid = uint(lastID)
	return
}

func (m *mysqlAccountRepository) DeleteAccount(ctx context.Context, account_no string) (err error) {
	// query := "DELETE FROM banking.account WHERE id = ?"

	query := `UPDATE banking.accounts set updated_at=? , is_deleted = 1 WHERE account_no = ?`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, time.Now(), account_no)
	if err != nil {
		return
	}

	rowsAfected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAfected != 1 {
		err = fmt.Errorf("weird  Behavior. Total Affected: %d", rowsAfected)
		return
	}

	if rowsAfected == 1 {
		err = fmt.Errorf("Delete completed")
		return
	}

	return
}

func (m *mysqlAccountRepository) UpdateAccount(ctx context.Context, ar *domain.Account) (err error) {
	query := `UPDATE banking.accounts set balance=?, updated_at=? WHERE account_no = ?`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	*ar.UpdatedAt = time.Now()

	res, err := stmt.ExecContext(ctx, ar.Balance, ar.UpdatedAt, ar.AccountNo)
	if err != nil {
		return
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {
		err = fmt.Errorf("weird  Behavior. Total Affected: %d", affect)
		return
	}

	cacheKey := fmt.Sprintf("account_no: %s", ar.AccountNo)

	if affect == 1 {
		err := m.redis.Del(cacheKey).Err()
		if err != nil {
			fmt.Printf("Error clearing key '%s': %v\n", cacheKey, err)
			return err
		}
		return err
	}

	return
}

// tx, err := m.conn.BeginTx(ctx, nil)
// 	if err != nil {
// 		return domain.Account{}, err
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 			return
// 		}
// 		tx.Commit()
// 	}()
