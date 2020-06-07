package internal

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const loggerKey = "logger"

var logger = zap.New(zapcore.NewCore(
	zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	zapcore.Lock(os.Stdout),
	zap.NewAtomicLevel(),
))

func WithLogger(cxt context.Context) context.Context {
	return context.WithValue(cxt, loggerKey, logger)
}
