package domain

import (
	"context"
	"time"
)

// Account is representing the Account data struct
type Account struct {
	AccountNo string     `json:"account_no,omitempty"`
	Uuid      string     `json:"uuid,omitempty"`
	Name      string     `json:"name,omitempty"`
	Email     string     `json:"email,omitempty"`
	Tel       string     `json:"tel,omitempty"`
	Balance   float64    `json:"balance"`
	Bank      string     `json:"bank,omitempty"`
	Status    string     `json:"status,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	IsClosed  int        `json:"is_closed,omitempty"`
}

type CountAccount struct {
	Status string `json:"status,"`
	Count  int    `json:"count"`
}

// AccountUsecase represent the account's usecases
type AccountUsecase interface {
	GetAllAccount(ctx context.Context, cursor string, num int64) ([]Account, string, error)
	GetAccountByAccountNo(ctx context.Context, account_no string) (*Account, error)
	UpdateAccount(ctx context.Context, ar *Account) error
	RegisterAccount(context.Context, *Account) error
	DeleteAccount(ctx context.Context, account_no string) error
	GetCountAccount(ctx context.Context) (map[string]int, error)
	// CalNewBalance(ctx context.Context, ar *Account, tr *Transaction) error
	ValidateAccount(ctx context.Context, ar *Account) error
	GetAllAccountByUuid(c context.Context, uuid string) (res *[]Account, err error)
	GetDailyLimit(c context.Context, account_no string) (float64, error)
	GetSumDailyTransaction(c context.Context, account_no string) (float64, error)
	SelectBank(lastDigit string) (bank string)
}

// AccountRepository represent the account's repository contract
type AccountRepository interface {
	GetAllAccount(ctx context.Context, cursor string, num int64) (res []Account, nextCursor string, err error)
	GetAccountFromRedisByAccountNo(ctx context.Context, account_no string) (*Account, error)
	GetAccountByAccountNo(ctx context.Context, account_no string) (*Account, error)
	UpdateAccount(ctx context.Context, ar *Account) error
	RegisterAccount(ctx context.Context, a *Account) error
	GetCountAccountByStatus(ctx context.Context) (result map[string]int, err error)
	DeleteAccount(ctx context.Context, account_no string) error
	GetAllAccountByUuid(ctx context.Context, uuid string) (res *[]Account, err error)
}
