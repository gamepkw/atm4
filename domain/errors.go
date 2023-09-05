package domain

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error Code: %d, Message: %s", e.Code, e.Message)
}

var (
	// ErrInternalServerError will throw if any the Internal Server Error happen
	ErrInternalServerError = errors.New("internal Server Error")
	// ErrNotFound will throw if the requested item is not exists
	ErrNotFound          = errors.New("Account not found")
	ErrResipientNotFound = errors.New("Resipient Account not found")
	ErrAccDeleted        = errors.New("Resipient Account closed")
	// ErrConflict will throw if the current action already exists
	ErrConflict = errors.New("your Item already exist")
	// ErrBadParamInput will throw if the given request-body or params is not valid
	ErrBadParamInput                   = errors.New("given Param is not valid")
	ErrInsufficientBalance             = errors.New("insufficient balance")
	ErrMinimumDeposit                  = errors.New("minimum for deposit is 100")
	ErrExceedLimitAmountPerTransaction = errors.New("exceed limit amount per transaction")
	ErrDuplicateUUID                   = errors.New("User already exists")
	ErrInvalidPassword                 = errors.New("Invalid password")
	ErrWrongPassword                   = &Error{Code: 1002, Message: "Wrong password"}
	ErrUserNotFound                    = &Error{Code: 1001, Message: "User not found"}
	ErrSetPin                          = errors.New("Can not set pin")
)
