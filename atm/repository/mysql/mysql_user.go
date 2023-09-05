package mysql

import (
	"context"
	"database/sql"
	"log"

	"main/domain"
	"time"

	"github.com/go-sql-driver/mysql"
)

type mysqlUserRepository struct {
	conn *sql.DB
}

// NewMysqlUserRepository will create an object that represent the user.Repository interface
func NewMysqlUserRepository(conn *sql.DB) domain.UserRepository {
	return &mysqlUserRepository{
		conn: conn,
	}
}

func (m *mysqlUserRepository) RegisterUser(ctx context.Context, u *domain.User) (res domain.UserResponse, err error) {
	query := `INSERT INTO banking.users SET uuid=?, hashed_password=?, created_at=?`
	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		if isDuplicateEntryError(err) {
			return res, domain.ErrDuplicateUUID
		}
	}

	_, err = stmt.ExecContext(ctx, u.Tel, u.Password, time.Now())
	if err != nil {
		return
	}
	res.UUID = u.Tel
	res.HashedPassword = u.Password

	return
}

func (m *mysqlUserRepository) SetPassword(ctx context.Context, u *domain.UpdatePassword) (res domain.UserResponse, err error) {
	query := `UPDATE banking.users SET hashed_password=?, updated_at=? WHERE uuid=?`
	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	_, err = stmt.ExecContext(ctx, u.NewPassword, time.Now(), u.Tel)
	if err != nil {
		return
	}
	res.UUID = u.Tel
	res.HashedPassword = u.NewPassword

	return
}

// func (m *mysqlUserRepository) UpdateVerifiedUser(ctx context.Context, u *domain.SetPassword) (res domain.UserResponse, err error) {
// 	query := `UPDATE banking.users SET is_verified=?, updated_at=? WHERE uuid=?`
// 	stmt, err := m.conn.PrepareContext(ctx, query)
// 	if err != nil {
// 		return
// 	}

// 	_, err = stmt.ExecContext(ctx, u.IsVerify, time.Now(), u.Tel)
// 	if err != nil {
// 		return
// 	}

// 	return
// }

// func (m *mysqlUserRepository) setOtpSecretInvalid(ctx context.Context, uuid string) (err error) {
// 	markInvalidStatus := "invalid"
// 	query := `UPDATE banking.users_otp SET status=? WHERE uuid=?`
// 	stmt, err := m.conn.PrepareContext(ctx, query)
// 	if err != nil {
// 		return
// 	}

// 	_, err = stmt.ExecContext(ctx, markInvalidStatus, uuid)
// 	if err != nil {
// 		return
// 	}
// 	return err
// }

func isDuplicateEntryError(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	return ok && mysqlErr.Number == 1062
}

func (m *mysqlUserRepository) GetHashedPasswordByUUID(ctx context.Context, uuid string) (res *domain.UserResponse, err error) {

	userResponse, err := m.fetchPasswordFromDatabase(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &userResponse, err

}

func (m *mysqlUserRepository) fetchPasswordFromDatabase(ctx context.Context, uuid string) (res domain.UserResponse, err error) {
	query := `SELECT hashed_password FROM banking.users WHERE uuid = ?`

	rows, err := m.conn.Query(query, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	found := false

	for rows.Next() {
		found = true
		err := rows.Scan(&res.HashedPassword)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if !found {
		return res, domain.ErrUserNotFound
	}

	return
}

func (m *mysqlUserRepository) GetHashedPinByUUID(ctx context.Context, uuid string) (res *domain.Pin, err error) {
	pinDB, err := m.fetchPinFromDatabase(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &pinDB, err

}

func (m *mysqlUserRepository) fetchPinFromDatabase(ctx context.Context, uuid string) (res domain.Pin, err error) {
	query := `SELECT hashed_pin FROM banking.users WHERE uuid = ?`

	rows, err := m.conn.Query(query, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	found := false

	for rows.Next() {
		found = true
		err := rows.Scan(&res.Pin)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if !found {
		return res, domain.ErrUserNotFound
	}

	return
}

func (m *mysqlUserRepository) SetUpPin(ctx context.Context, u *domain.Pin) (err error) {
	query := `UPDATE banking.users SET hashed_pin=? WHERE uuid=?`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	res, err := stmt.ExecContext(ctx, u.Pin, u.Tel)
	if err != nil {
		return err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {

		return domain.ErrSetPin
	}

	return
}

func (m *mysqlUserRepository) SetNewPin(ctx context.Context, u *domain.SetNewPin) (err error) {
	query := `UPDATE banking.users SET hashed_pin=? WHERE uuid=?`

	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	res, err := stmt.ExecContext(ctx, u.NewPin, u.Tel)
	if err != nil {
		return err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {
		return err
	}
	return
}
