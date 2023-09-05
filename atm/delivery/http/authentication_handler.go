package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"main/domain"
)

// ResponseError represent the response error struct

// UserHandler  represent the httphandler for user
type AuthenticationHandler struct {
	AuthUsecase domain.AuthenticationUsecase
}

// NewUserHandler will initialize the users/ resources endpoint
func NewAuthenticationHandler(e *echo.Echo, auths domain.AuthenticationUsecase) {
	handler := &AuthenticationHandler{
		AuthUsecase: auths,
	}

	e.POST("/users/send-otp", handler.SendOtp)
	e.POST("/users/verify-otp", handler.ValidateOtp)

}

func (auth *AuthenticationHandler) SendOtp(c echo.Context) (err error) {

	var set domain.UpdatePassword
	if err = c.Bind(&set); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	ctx := c.Request().Context()

	if err = auth.AuthUsecase.SendOtp(ctx, set.Tel); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response{Message: "Send otp successfully", Body: nil})

}

func (auth *AuthenticationHandler) ValidateOtp(c echo.Context) (err error) {
	var set domain.UpdatePassword

	if err = c.Bind(&set); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	ctx := c.Request().Context()

	if !auth.AuthUsecase.ValidateOtp(ctx, set.Tel, set.Otp) {
		return c.JSON(http.StatusBadRequest, Response{Message: "Otp is invalid", Body: nil})
	}

	return c.JSON(http.StatusOK, Response{Message: "Otp is valid", Body: nil})

}
