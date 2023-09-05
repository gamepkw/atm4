package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"

	"main/atm/delivery/http/middleware"
	"main/domain"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AccountHandler  represent the httphandler for account
type AccountHandler struct {
	AUsecase domain.AccountUsecase
}

type AccountResponse struct {
	Message string          `json:"message"`
	Body    *domain.Account `json:"body,omitempty"`
}

// NewAccountHandler will initialize the accounts/ resources endpoint
func NewAccountHandler(e *echo.Echo, us domain.AccountUsecase) {
	handler := &AccountHandler{
		AUsecase: us,
	}
	restrictedGroup := e.Group("/users/accounts")
	restrictedGroup.Use(middleware.CustomJWTMiddleware)

	e.GET("/accounts", handler.GetAllAccount)
	// e.POST("/accounts/register", handler.RegisterAccount)
	e.GET("/accounts/:account_no", handler.GetAccountByAccountNo)
	e.PUT("/accounts/:account_no", handler.UpdateAccount)
	e.PUT("/accounts/:account_no", handler.CloseAccount)
	e.GET("/accounts/get-count-by-status", handler.GetCountAccount)

	restrictedGroup.GET("/get-all-account", handler.GetAllAccountByUuid)
	restrictedGroup.POST("/register", handler.RegisterAccount)
}

// FetchAccount will fetch the account based on given params
func (a *AccountHandler) GetAllAccount(c echo.Context) error {
	numS := c.QueryParam("num")
	num, _ := strconv.Atoi(numS)
	cursor := c.QueryParam("cursor")

	ctx := c.Request().Context()
	// listAr, nextCursor, err := a.AUsecase.Fetch(ctx, cursor, int64(num))
	listAr, nextCursor, err := a.AUsecase.GetAllAccount(ctx, cursor, int64(num))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	c.Response().Header().Set(`X-Cursor`, nextCursor)
	return c.JSON(http.StatusOK, listAr)
}

func (a *AccountHandler) GetAllAccountByUuid(c echo.Context) error {
	uuid := c.Get("tel").(string)
	ctx := c.Request().Context()

	accounts, err := a.AUsecase.GetAllAccountByUuid(ctx, uuid)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, accounts)
}

func (a *AccountHandler) GetAccountByAccountNo(c echo.Context) error {
	account_no := c.Param("account_no")
	ctx := c.Request().Context()

	account, err := a.AUsecase.GetAccountByAccountNo(ctx, account_no)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, account)
}

func (a *AccountHandler) GetCountAccount(c echo.Context) error {
	ctx := c.Request().Context()

	list, err := a.AUsecase.GetCountAccount(ctx)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, list)
}

// Store will store the account by given request body
func (a *AccountHandler) RegisterAccount(c echo.Context) (err error) {

	uuid := c.Get("tel").(string)

	var account domain.Account

	if err = c.Bind(&account); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	account.Uuid = uuid

	var ok bool
	if ok, err = isRequestValid(&account); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	if err = a.AUsecase.RegisterAccount(ctx, &account); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	res, err := a.AUsecase.GetAccountByAccountNo(ctx, account.AccountNo)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, res)
}

func (a *AccountHandler) UpdateAccount(c echo.Context) (err error) {
	var account domain.Account
	err = c.Bind(&account)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	var ok bool
	if ok, err = isRequestValid(&account); !ok {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	err = a.AUsecase.UpdateAccount(ctx, &account)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, AccountResponse{Message: "Update account successfully", Body: &account})
}

// Delete will delete account by given param
func (a *AccountHandler) CloseAccount(c echo.Context) error {
	idP, err := strconv.Atoi(c.Param("sender"))
	if err != nil {
		return c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	sender := string(idP)
	ctx := c.Request().Context()

	err = a.AUsecase.DeleteAccount(ctx, sender)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func isRequestValid(m *domain.Account) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}
