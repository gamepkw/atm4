package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mrwaggel/golimiter"
)

// GoMiddleware represent the data-struct for middleware
const limitPerMinute = 100

var limiter = golimiter.New(limitPerMinute, time.Minute)

func (m *GoMiddleware) RateLimitMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if the IP is limited
		if limiter.IsLimited(c.RealIP()) {
			return c.String(http.StatusTooManyRequests, "429: Too many requests")
		}
		// Increment the value for the IP
		limiter.Increment(c.RealIP())
		// Continue default operation of Echo
		return next(c)
	}
}

const limitTransactionPerMinute = 100

var limiterForTransaction = golimiter.New(limitTransactionPerMinute, time.Minute)

func (m *GoMiddleware) RateLimitMiddlewareForTransaction(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if the IP is limited
		if limiterForTransaction.IsLimited(c.RealIP()) {
			return c.String(http.StatusTooManyRequests, "429: Too many requests")
		}
		// Increment the value for the IP
		limiterForTransaction.Increment(c.RealIP())
		// Continue default operation of Echo
		return next(c)
	}
}
