package mysql

import (
	"context"
	"database/sql"
	"log"

	"main/domain"
	"time"

	"github.com/go-redis/redis"
)

type mysqlAuthenticationRepository struct {
	conn  *sql.DB
	redis *redis.Client
}

func NewMysqlAuthenticationRepository(conn *sql.DB, redis *redis.Client) domain.AuthenticationRepository {
	return &mysqlAuthenticationRepository{
		conn:  conn,
		redis: redis,
	}
}

func (m *mysqlAuthenticationRepository) GetOtpSecret(ctx context.Context, uuid string) (secretKey string, expiredAt time.Time, err error) {
	query := `SELECT secret, expired_at FROM banking.users_otp WHERE uuid = ?`

	rows, err := m.conn.Query(query, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&secretKey, &expiredAt)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return secretKey, expiredAt, nil
}

func (m *mysqlAuthenticationRepository) SaveOtpSecret(ctx context.Context, uuid string, secretKey string) (err error) {
	query := `INSERT INTO banking.users_otp SET uuid=?, secret=?, created_at=?, expired_at=?`
	stmt, err := m.conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	_, err = stmt.ExecContext(ctx, uuid, secretKey, time.Now(), time.Now().Add(180*time.Second))
	if err != nil {
		return
	}
	return err
}
