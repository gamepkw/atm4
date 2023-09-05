package domain

import (
	"context"
)

type User struct {
	Tel      string `json:"tel"`
	Password string `json:"password"`
}

type UserResponse struct {
	UUID           string `json:"uuid"`
	HashedPassword string `json:"hashed_password"`
}

type Pin struct {
	Tel string `json:"tel"`
	Pin string `json:"pin"`
}

type SetNewPin struct {
	Tel    string `json:"tel"`
	Pin    string `json:"pin"`
	NewPin string `json:"new_pin"`
}

type UpdatePassword struct {
	Tel         string `json:"tel"`
	Otp         string `json:"otp"`
	Password    string `json:"current_password"`
	NewPassword string `json:"new_password"`
}

type UserUsecase interface {
	RegisterUser(context.Context, *User) (UserResponse, error)
	Login(c context.Context, u *User) (token string, err error)
	SetUpPin(c context.Context, u *Pin) (err error)
	SetNewPin(c context.Context, u *SetNewPin) (err error)
	SetNewPassword(c context.Context, u *UpdatePassword) (res UserResponse, err error)
	ResetPassword(c context.Context, u *UpdatePassword) (res UserResponse, err error)
	ValidatePin(c context.Context, uuid string, pin string) bool
}

// UserRepository represent the user's repository contract
type UserRepository interface {
	RegisterUser(ctx context.Context, u *User) (res UserResponse, err error)
	GetHashedPasswordByUUID(ctx context.Context, uuid string) (res *UserResponse, err error)
	SetUpPin(ctx context.Context, u *Pin) (err error)
	SetNewPin(ctx context.Context, u *SetNewPin) (err error)
	GetHashedPinByUUID(ctx context.Context, uuid string) (res *Pin, err error)
	SetPassword(ctx context.Context, u *UpdatePassword) (res UserResponse, err error)
}
