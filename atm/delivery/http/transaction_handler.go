package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"main/atm/delivery/http/middleware"
	"main/atm/utils"
	"main/logger"

	"main/domain"

	"github.com/go-redis/redis"
)

// ResponseError represent the response error struct

// TransactionHandler  represent the httphandler for transaction
type TransactionHandler struct {
	TrUsecase domain.TransactionUsecase
	AcUsecase domain.AccountUsecase
	redis     *redis.Client
}

type TransactionResponse struct {
	Message string              `json:"message"`
	Body    *domain.Transaction `json:"body,omitempty"`
}

// func transactionapiGroup(c echo.Context) error {
// 	user := c.Get("user").(*jwt.Token)
// 	claims := user.Claims.(jwt.MapClaims)
// 	username := claims["username"].(string)
// 	return c.String(http.StatusOK, fmt.Sprintf("Hello, %s! You are in the restricted group.", username))
// }

// NewTransactionHandler will initialize the transactions/ resources endpoint
func NewTransactionHandler(e *echo.Echo, us domain.TransactionUsecase, redis *redis.Client) {
	handler := &TransactionHandler{
		TrUsecase: us,
		redis:     redis,
	}

	middL := middleware.InitMiddleware()

	transactionapiGroup := e.Group("/transaction", middL.RateLimitMiddlewareForTransaction)

	// e.GET("/transactions", handler.GetAllTransaction)
	transactionapiGroup.POST("/deposit", handler.Deposit)
	transactionapiGroup.POST("/withdraw", handler.Withdraw)
	transactionapiGroup.POST("/transfer", handler.Transfer)
	transactionapiGroup.POST("/schedule", handler.ScheduledTransaction)
}

var transferRequest = "POST /transaction/transfer"

func (a *TransactionHandler) Deposit(c echo.Context) (err error) {
	var transaction domain.Transaction
	transaction.SubmittedAt = time.Now()

	if err = c.Bind(&transaction); err != nil {

		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if transaction.Type != "deposit" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	ctx := c.Request().Context()

	if err = a.TrUsecase.Deposit(ctx, &transaction); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, TransactionResponse{Message: "Deposit successfully", Body: &transaction})
}

func (a *TransactionHandler) Withdraw(c echo.Context) (err error) {
	var transaction domain.Transaction

	ctx := c.Request().Context()

	if err = c.Bind(&transaction); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if transaction.Type != "withdraw" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	// ctx := c.Request().Context()

	transaction.SubmittedAt = time.Now()

	if err = a.TrUsecase.Withdraw(ctx, &transaction); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, TransactionResponse{Message: "Withdraw successfully", Body: &transaction})
}

func (a *TransactionHandler) Transfer(c echo.Context) error {
	// logger.Info(fmt.Sprintf("%s: start...", transferRequest), c.Request())
	var transaction domain.Transaction

	requestBody := utils.UnmarshalRequestBody(c.Request())

	if err := c.Bind(&transaction); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if transaction.Type != "transfer" {
		logger.Error(fmt.Sprintf("%s: Invalid request \n %s", transferRequest, requestBody), c.Request())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	if transaction.Account.AccountNo == transaction.Receiver.AccountNo {
		logger.Error(fmt.Sprintf("%s: Can not transfer to the same account \n %s", transferRequest, requestBody), c.Request())
		return echo.NewHTTPError(http.StatusBadRequest, "Can not transfer to the same account")
	}

	if transaction.Amount <= 0 {
		logger.Error(fmt.Sprintf("%s: Transfer amount must be positive \n %s", transferRequest, requestBody), c.Request())
		return echo.NewHTTPError(http.StatusBadRequest, "Transfer amount must be positive")
	}

	ctx := c.Request().Context()
	transaction.SubmittedAt = time.Now()

	if err := a.TrUsecase.Transfer(ctx, &transaction); err != nil {
		logger.Error(fmt.Sprintf("%s %s \n %s", transferRequest, err.Error(), requestBody), c.Request())
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	// logger.Info(fmt.Sprintf("%s: stop...", transferRequest), c.Request())
	return c.JSON(http.StatusCreated, TransactionResponse{Message: "Transfer successfully", Body: &transaction})
}

func (a *TransactionHandler) ScheduledTransaction(c echo.Context) error {
	var transaction domain.ScheduledTransaction

	if err := c.Bind(&transaction); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if transaction.Type != "transfer" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	if transaction.Account.AccountNo == transaction.Receiver.AccountNo {
		return echo.NewHTTPError(http.StatusBadRequest, "Can not transfer to the same account")
	}

	if transaction.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Transfer amount must be positive")
	}

	ctx := c.Request().Context()
	transaction.SubmittedAt = time.Now()

	if err := a.TrUsecase.SaveScheduledTransaction(ctx, &transaction); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, TransactionResponse{Message: "Set scheduled transfer successfully", Body: nil})
}
