package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"main/atm/delivery/http/middleware"
	"main/atm/utils"
	"main/domain"

	"github.com/sirupsen/logrus"
)

// ResponseError represent the response error struct

// UserHandler  represent the httphandler for user
type UserHandler struct {
	UUsecase    domain.UserUsecase
	AuthUsecase domain.AuthenticationUsecase
}

type Response struct {
	Message string               `json:"message"`
	Body    *domain.UserResponse `json:"body,omitempty"`
}

type LoginResponse struct {
	Message string               `json:"message"`
	Token   string               `json:"token"`
	Body    *domain.UserResponse `json:"body,omitempty"`
}

// NewUserHandler will initialize the users/ resources endpoint
func NewUserHandler(e *echo.Echo, us domain.UserUsecase, auths domain.AuthenticationUsecase) {
	handler := &UserHandler{
		UUsecase:    us,
		AuthUsecase: auths,
	}

	restrictedGroup := e.Group("/users/pin")
	restrictedGroup.Use(middleware.CustomJWTMiddleware)

	e.POST("/users/register", handler.RegisterUser)
	e.POST("/users/login", handler.Login)
	e.POST("/users/set-new-password", handler.ResetPassword)
	restrictedGroup.PUT("/set-pin", handler.SetUpPin)
	restrictedGroup.PUT("/set-new-pin", handler.SetNewPin)
	restrictedGroup.POST("/verify-pin", handler.ValidatePin)
}

var TimestampFormat = "2006-01-02 15:04:05"

func (a *UserHandler) RegisterUser(c echo.Context) (err error) {
	var user domain.User

	if err = c.Bind(&user); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if user.Tel == "" || len(user.Tel) != 10 {
		logrus.Errorf("[RegisterUser] Invalid Tel")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Tel")
	}

	if user.Password == "" {
		logrus.Errorf("[RegisterUser] Empty Password")
		return echo.NewHTTPError(http.StatusBadRequest, "Empty Password")
	}

	ctx := c.Request().Context()
	res, err := a.UUsecase.RegisterUser(ctx, &user)
	if err != nil {
		logrus.Errorf("[RegisterUser] %s", err.Error())
		return c.JSON(getStatusCode(err), ResponseError{Code: "1000", Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, Response{Message: "Register successful", Body: &res})
}

func (a *UserHandler) Login(c echo.Context) (err error) {
	// time.Sleep(3 * time.Second)
	var user domain.User

	if err = c.Bind(&user); err != nil {
		logrus.Errorf("[Login] %s", err.Error())
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if user.Tel == "" || len(user.Tel) != 10 {
		logrus.Errorf("[Login] Invalid Tel")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Tel")
	}

	if user.Password == "" {
		logrus.Errorf("[Login] Empty Password")
		return echo.NewHTTPError(http.StatusBadRequest, "Empty Password")
	}

	ctx := c.Request().Context()

	token, err := a.UUsecase.Login(ctx, &user)
	if err != nil {
		logrus.Errorf("[Login] %s", err)
		return c.JSON(getStatusCode(err), err)
	}

	// token, err := middleware.GenerateJWTToken(user.Tel, 1*time.Hour)
	// if err != nil {
	// 	return err
	// }

	return c.JSON(http.StatusOK, LoginResponse{Message: "Login successful", Token: token, Body: nil})

}

func (a *UserHandler) ResetPassword(c echo.Context) (err error) {
	var user domain.UpdatePassword

	if err = c.Bind(&user); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if user.Tel == "" || len(user.Tel) != 10 {
		logrus.Errorf("[ResetPassword] Invalid Tel")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Tel")
	}

	if user.NewPassword == "" {
		logrus.Errorf("[ResetPassword] Empty Password")
		return echo.NewHTTPError(http.StatusBadRequest, "Empty Password")
	}

	if user.Password == user.NewPassword {
		logrus.Errorf("[ResetPassword] New password can not be the cuurent password")
		return echo.NewHTTPError(http.StatusBadRequest, "New password can not be the cuurent password")
	}

	ctx := c.Request().Context()

	res, err := a.UUsecase.ResetPassword(ctx, &user)
	if err != nil {
		logrus.Errorf("[ResetPassword] %s", err)
		return c.JSON(getStatusCode(err), err)
	}

	return c.JSON(http.StatusOK, Response{Message: "Set new password successful", Body: &res})

}

func (a *UserHandler) ValidatePin(c echo.Context) (err error) {
	var pin domain.Pin

	uuid := c.Get("tel").(string)

	if err = c.Bind(&pin); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	pin.Tel = uuid

	ctx := c.Request().Context()

	if !a.UUsecase.ValidatePin(ctx, pin.Tel, pin.Pin) {
		return c.JSON(http.StatusBadRequest, Response{Message: "Pin is incorrect", Body: nil})
	}

	return c.JSON(http.StatusOK, Response{Message: "Pin is valid", Body: nil})

}

func (a *UserHandler) VerifyUser(c echo.Context) (err error) {
	var set domain.UpdatePassword

	if err = c.Bind(&set); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	ctx := c.Request().Context()

	if !a.AuthUsecase.ValidateOtp(ctx, set.Tel, set.Otp) {
		return c.JSON(http.StatusBadRequest, Response{Message: "Otp is invalid", Body: nil})
	}

	return c.JSON(http.StatusOK, Response{Message: "Otp is valid", Body: nil})

}

func (a *UserHandler) SetUpPin(c echo.Context) (err error) {
	var pin domain.Pin

	expectedTel := c.Get("tel").(string)

	if err = c.Bind(&pin); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if err := utils.EncodeBase64(&pin.Tel); err != nil {
		return err
	}

	if pin.Tel != expectedTel {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Tel mismatch"})
	}

	if pin.Pin == "" || len(pin.Pin) != 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Pin")
	}

	ctx := c.Request().Context()

	if err = a.UUsecase.SetUpPin(ctx, &pin); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, Response{Message: "Set pin successfully", Body: nil})
}

func (a *UserHandler) SetNewPin(c echo.Context) (err error) {
	var pin domain.SetNewPin

	expectedTel := c.Get("tel").(string)

	if err = c.Bind(&pin); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if err := utils.EncodeBase64(&pin.Tel); err != nil {
		return err
	}

	if pin.Tel != expectedTel {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "Unauthorized"})
	}

	if pin.Pin == "" || len(pin.Pin) != 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Pin")
	}

	if pin.NewPin == "" || len(pin.NewPin) != 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid New Pin")
	}
	if pin.NewPin == pin.Pin {
		return echo.NewHTTPError(http.StatusBadRequest, "New pin same as old pin")
	}

	ctx := c.Request().Context()

	if err = a.UUsecase.SetNewPin(ctx, &pin); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: "Set new pin failed"})
	}

	return c.JSON(http.StatusOK, Response{Message: "Set new pin successfully", Body: nil})
}
