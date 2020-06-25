package internal

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger = zap.New(zapcore.NewCore(
	zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	zapcore.Lock(os.Stdout),
	zap.NewAtomicLevel(),
)).Sugar()
