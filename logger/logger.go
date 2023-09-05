package logger

import (
	"context"
	"net/http"

	"cloud.google.com/go/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel string

var globalLogger *zap.Logger
var gcpLogger *logging.Logger

func init() {
	cfg := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel), // Default to Info level
		OutputPaths: []string{"stdout"},
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	globalLogger = logger

	ctx := context.Background()
	client, err := logging.NewClient(ctx, "banking-atm-397503")
	if err != nil {
		panic("Failed to create GCP logger client: " + err.Error())
	}

	gcpLogger = client.Logger("atm")
}

func SetLogLevel(logLevel LogLevel) {
	// ... (same as before)
}

func logStructured(loggerFn func(msg string, fields ...zap.Field), level logging.Severity, message string) {
	loggerFn("", zap.String("message", message))
	gcpLogger.Log(logging.Entry{Severity: level, Payload: message})
}

func Info(message string, fields ...zapcore.Field) {
	globalLogger.Info(message)
	gcpLogger.Log(logging.Entry{Severity: logging.Info, Payload: message})
}

func Debug(message string) {
	globalLogger.Debug(message)
	gcpLogger.Log(logging.Entry{Severity: logging.Debug, Payload: message})
}

func Warning(message string, request *http.Request, fields ...zapcore.Field) {
	globalLogger.Warn(message)
	httpRequest := &logging.HTTPRequest{
		Request: request,
	}
	gcpLogger.Log(logging.Entry{Severity: logging.Warning, Payload: message, HTTPRequest: httpRequest})
}

func Error(message string, request *http.Request, fields ...zapcore.Field) {
	globalLogger.Error(message, fields...)
	httpRequest := &logging.HTTPRequest{
		Request: request,
	}
	gcpLogger.Log(logging.Entry{Severity: logging.Error, Payload: message, HTTPRequest: httpRequest})
}
